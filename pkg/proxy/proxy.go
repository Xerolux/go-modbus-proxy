package proxy

import (
	"context"
	"fmt"
	"io"
	"modbusproxy/pkg/devices"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/modbus"
	"modbusproxy/pkg/pool"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ProxyInstance represents a running proxy.
type ProxyInstance struct {
	ID         string
	Name       string
	ListenAddr string
	TargetAddr string

	// Config
	MaxReadSize       int
	ConnectionTimeout time.Duration
	ReadTimeout       time.Duration
	MaxRetries        int

	listener net.Listener
	connPool *pool.Pool
	targetMu sync.Mutex

	log           *logger.Logger
	deviceTracker *devices.Tracker
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	Stats Stats
}

type Stats struct {
	Uptime    time.Duration
	LastStart time.Time
	Requests  atomic.Int64
	Errors    atomic.Int64
	Status    string // "Running", "Stopped", "Error"
}

// NewProxyInstance creates a new proxy.
func NewProxyInstance(id, name, listen, target string, maxReadSize int, l *logger.Logger, tracker *devices.Tracker) *ProxyInstance {
	return &ProxyInstance{
		ID:                id,
		Name:              name,
		ListenAddr:        listen,
		TargetAddr:        target,
		MaxReadSize:       maxReadSize,
		ConnectionTimeout: 5 * time.Second, // Default
		ReadTimeout:       5 * time.Second, // Default
		MaxRetries:        3,               // Default
		log:               l,
		deviceTracker:     tracker,
		Stats:             Stats{Status: "Stopped"},
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
		p.log.Error(p.ID, fmt.Sprintf("Failed to listen: %v", err))
		return err
	}
	p.listener = l

	// Create connection pool for target with optimized settings
	poolCfg := pool.Config{
		InitialSize:    2, // Optimized: Better initial capacity
		MaxSize:        20, // Optimized: Increased for higher concurrency
		MaxIdleTime:    10 * time.Minute, // Optimized: Longer idle time for reusability
		AcquireTimeout: p.ConnectionTimeout,
		Dialer: func(ctx context.Context) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   p.ConnectionTimeout,
				KeepAlive: 30 * time.Second, // Optimized: TCP keep-alive
			}
			return d.DialContext(ctx, "tcp", p.TargetAddr)
		},
	}

	p.connPool, err = pool.NewPool(poolCfg)
	if err != nil {
		p.listener.Close()
		p.Stats.Status = "Error"
		p.log.Error(p.ID, fmt.Sprintf("Failed to create connection pool: %v", err))
		return err
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.Stats.Status = "Running"
	p.Stats.LastStart = time.Now()

	p.log.Info(p.ID, fmt.Sprintf("Started proxy listening on %s -> %s", p.ListenAddr, p.TargetAddr))

	p.wg.Add(1)
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
		p.listener.Close()
	}
	if p.connPool != nil {
		p.connPool.Close()
	}

	// Wait for all goroutines to finish
	p.wg.Wait()

	p.Stats.Status = "Stopped"
}

func (p *ProxyInstance) acceptLoop() {
	defer p.wg.Done()

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.ctx.Done():
				return // Normal shutdown
			default:
				p.log.Error(p.ID, fmt.Sprintf("Accept error: %v", err))
				// Backoff slightly to avoid spinning
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		p.wg.Add(1)
		go p.handleClient(conn)
	}
}

func (p *ProxyInstance) handleClient(clientConn net.Conn) {
	defer p.wg.Done()
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

		if err := clientConn.SetReadDeadline(time.Now().Add(p.ReadTimeout)); err != nil {
			p.log.Error(p.ID, fmt.Sprintf("SetReadDeadline failed: %v", err))
			return
		}
		reqFrame, err := modbus.ReadFrame(clientConn)
		if err != nil {
			if err != io.EOF {
				p.log.Info(p.ID, fmt.Sprintf("Client read error: %v", err))
			}
			return
		}

		var respFrame []byte
		var errFwd error

		// Check if splitting is needed
		if p.MaxReadSize > 0 && modbus.IsReadRequest(reqFrame) {
			respFrame, errFwd = p.handleSplitRead(reqFrame)
		} else {
			respFrame, errFwd = p.forwardRequest(reqFrame)
		}

		if errFwd != nil {
			p.log.Error(p.ID, fmt.Sprintf("Forward error: %v", errFwd))
			p.Stats.Errors.Add(1)
			return
		}
		p.Stats.Requests.Add(1)

		if _, err := clientConn.Write(respFrame); err != nil {
			p.log.Error(p.ID, fmt.Sprintf("Write response error: %v", err))
			return
		}
	}
}

func (p *ProxyInstance) handleSplitRead(reqFrame []byte) ([]byte, error) {
	txID, unitID, fc, startAddr, quantity, err := modbus.ParseReadRequest(reqFrame)
	if err != nil {
		// Malformed request, just forward it and let target fail or fail here
		return p.forwardRequest(reqFrame)
	}

	// If quantity is within limits, forward normally
	if int(quantity) <= p.MaxReadSize {
		return p.forwardRequest(reqFrame)
	}

	// Split logic
	expectedBytes := int(quantity) * 2 // 2 bytes per register
	aggregatedData := make([]byte, 0, expectedBytes)
	remaining := quantity
	currentAddr := startAddr

	// We need to use the target connection lock for the whole sequence?
	// If we don't, other requests might interleave.
	// Interleaving is generally fine for Modbus TCP unless the device is very sensitive.
	// But to be safe and atomic for the "Read Block", let's lock.
	// However, `forwardRequest` locks internally.
	// We can't lock `p.targetMu` here easily without refactoring `forwardRequest` or extracting the inner logic.

	// Refactoring to expose an internal unlocked forward function is better.
	p.targetMu.Lock()
	defer p.targetMu.Unlock()

	for remaining > 0 {
		chunkSize := uint16(p.MaxReadSize)
		if remaining < chunkSize {
			chunkSize = remaining
		}

		// Create sub-request
		// We use the same TxID? Or a counter?
		// The target doesn't care about TxID usually, but good to be consistent.
		subReq := modbus.CreateReadRequest(0, unitID, fc, currentAddr, chunkSize)

		// Send using internal (unlocked) forward
		subResp, err := p.forwardRequestLocked(subReq)
		if err != nil {
			return nil, err
		}

		// Parse response
		data, err := modbus.ParseReadResponse(subResp)
		if err != nil {
			// Check if it's an exception
			if len(subResp) > 7 && (subResp[7]&0x80) != 0 {
				// It is an exception. Return it immediately (with corrected TxID).
				// Fix TxID
				subResp[0] = byte(txID >> 8)
				subResp[1] = byte(txID)
				return subResp, nil
			}
			return nil, fmt.Errorf("split read failed: %v", err)
		}

		if len(data) != int(chunkSize)*2 {
			return nil, fmt.Errorf("split read received unexpected data length: got %d, want %d", len(data), chunkSize*2)
		}

		aggregatedData = append(aggregatedData, data...)
		remaining -= chunkSize
		currentAddr += chunkSize
	}

	// Construct final response
	respFrame := modbus.CreateReadResponse(txID, unitID, fc, aggregatedData)
	return respFrame, nil
}

// forwardRequestLocked assumes p.targetMu is already locked.
func (p *ProxyInstance) forwardRequestLocked(req []byte) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			time.Sleep(backoff)
		}

		// Get connection from pool
		conn, err := p.connPool.Get(p.ctx)
		if err != nil {
			lastErr = err
			continue
		}

		// Try write
		if err := conn.SetWriteDeadline(time.Now().Add(p.ConnectionTimeout)); err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		if _, err := conn.Write(req); err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		// Try read
		if err := conn.SetReadDeadline(time.Now().Add(p.ReadTimeout)); err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		resp, err := modbus.ReadFrame(conn)
		if err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		// Return connection to pool
		conn.Close()
		return resp, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", p.MaxRetries, lastErr)
}

func (p *ProxyInstance) forwardRequest(req []byte) ([]byte, error) {
	p.targetMu.Lock()
	defer p.targetMu.Unlock()
	return p.forwardRequestLocked(req)
}
