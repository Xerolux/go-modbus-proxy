package proxy

import (
	"context"
	"io"
	"modbusproxy/pkg/devices"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/modbus"
	"modbusproxy/pkg/pool"
	"net"
	"syscall"
	"time"
)

// PoolConfig holds connection pool configuration.
type PoolConfig struct {
	Size               int
	MinSize            int
	ConnTimeout        time.Duration
	MaxIdleTime        time.Duration
	KeepAlive          bool
	HealthCheckInterval time.Duration
}

// healthCheck verifies a connection is still alive by checking socket state.
func healthCheck(conn net.Conn) error {
	// Try to set a deadline to verify socket is still valid
	if err := conn.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return err
	}

	// For TCP connections, we can check if the socket is still open
	// by examining the underlying file descriptor
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Use SyscallConn to check socket state without actually reading/writing
		rawConn, err := tcpConn.SyscallConn()
		if err != nil {
			return err
		}

		var sysErr error
		err = rawConn.Control(func(fd uintptr) {
			// Try to get socket error state
			_, sysErr = syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_ERROR)
		})

		if err != nil {
			return err
		}
		if sysErr != nil {
			return sysErr
		}
	}

	// Reset deadline
	_ = conn.SetDeadline(time.Time{})
	return nil
}

// ProxyInstance represents a running proxy.
type ProxyInstance struct {
	ID         string
	Name       string
	ListenAddr string
	TargetAddr string

	listener      net.Listener
	pool          *pool.Pool
	poolConfig    PoolConfig

	log           *logger.Logger
	deviceTracker *devices.Tracker
	ctx           context.Context
	cancel        context.CancelFunc

	Stats         Stats
}

type Stats struct {
	Uptime       time.Duration
	LastStart    time.Time
	Requests     int64
	Errors       int64
	Status       string // "Running", "Stopped", "Error"
}

// NewProxyInstance creates a new proxy.
func NewProxyInstance(id, name, listen, target string, l *logger.Logger, tracker *devices.Tracker, poolCfg PoolConfig) *ProxyInstance {
	// Set defaults if not configured
	if poolCfg.Size == 0 {
		poolCfg.Size = 10
	}
	if poolCfg.MinSize == 0 {
		poolCfg.MinSize = 2
	}
	if poolCfg.ConnTimeout == 0 {
		poolCfg.ConnTimeout = 5 * time.Second
	}
	if poolCfg.MaxIdleTime == 0 {
		poolCfg.MaxIdleTime = 5 * time.Minute
	}
	if poolCfg.HealthCheckInterval == 0 {
		poolCfg.HealthCheckInterval = 60 * time.Second
	}

	return &ProxyInstance{
		ID:            id,
		Name:          name,
		ListenAddr:    listen,
		TargetAddr:    target,
		poolConfig:    poolCfg,
		log:           l,
		deviceTracker: tracker,
		Stats:         Stats{Status: "Stopped"},
	}
}

// Start starts the proxy.
func (p *ProxyInstance) Start() error {
	if p.Stats.Status == "Running" {
		return nil
	}

	l, err := net.Listen("tcp", p.ListenAddr)
	if err != nil {
		p.Stats.Status = "Error"
		p.log.Error(p.ID, "Failed to listen: "+err.Error())
		return err
	}
	p.listener = l

	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Initialize connection pool
	poolCfg := pool.Config{
		InitialSize:    p.poolConfig.MinSize,
		MaxSize:        p.poolConfig.Size,
		MaxIdleTime:    p.poolConfig.MaxIdleTime,
		AcquireTimeout: 30 * time.Second,
		Dialer: func(ctx context.Context) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   p.poolConfig.ConnTimeout,
				KeepAlive: -1, // Disable system keepalive, we handle it ourselves
			}
			if p.poolConfig.KeepAlive {
				d.KeepAlive = 30 * time.Second
			}
			return d.DialContext(ctx, "tcp", p.TargetAddr)
		},
		HealthChecker: healthCheck,
	}

	p.pool, err = pool.NewPool(poolCfg)
	if err != nil {
		p.listener.Close()
		p.Stats.Status = "Error"
		p.log.Error(p.ID, "Failed to create connection pool: "+err.Error())
		return err
	}

	p.Stats.Status = "Running"
	p.Stats.LastStart = time.Now()

	p.log.Info(p.ID, "Started proxy listening on "+p.ListenAddr+" -> "+p.TargetAddr)

	go p.acceptLoop()
	return nil
}

// Stop stops the proxy.
func (p *ProxyInstance) Stop() {
	if p.Stats.Status != "Running" {
		return
	}

	p.log.Info(p.ID, "Stopping proxy")
	p.cancel()
	if p.listener != nil {
		_ = p.listener.Close()
	}
	if p.pool != nil {
		_ = p.pool.Close()
	}
	p.Stats.Status = "Stopped"
}

func (p *ProxyInstance) acceptLoop() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.ctx.Done():
				return // Normal shutdown
			default:
				p.log.Error(p.ID, "Accept error: "+err.Error())
				// Backoff slightly to avoid spinning
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		go p.handleClient(conn)
	}
}

func (p *ProxyInstance) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	// Track the device connection
	if p.deviceTracker != nil {
		p.deviceTracker.TrackConnection(clientConn, p.ID)
	}

	for {
		// Check context
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		_ = clientConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		reqFrame, err := modbus.ReadFrame(clientConn)
		if err != nil {
			if err != io.EOF {
				p.log.Info(p.ID, "Client read error: "+err.Error())
			}
			return
		}

		respFrame, err := p.forwardRequest(reqFrame)
		if err != nil {
			p.log.Error(p.ID, "Forward error: "+err.Error())
			p.Stats.Errors++
			return
		}
		p.Stats.Requests++

		if _, err := clientConn.Write(respFrame); err != nil {
			p.log.Error(p.ID, "Write response error: "+err.Error())
			return
		}
	}
}

func (p *ProxyInstance) forwardRequest(req []byte) ([]byte, error) {
	// Get connection from pool with timeout
	ctx, cancel := context.WithTimeout(p.ctx, p.poolConfig.ConnTimeout)
	defer cancel()

	conn, err := p.pool.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // Returns connection to pool

	// Write request
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(req); err != nil {
		return nil, err
	}

	// Read response
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp, err := modbus.ReadFrame(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
