package proxy

import (
	"context"
	"io"
	"modbusproxy/pkg/devices"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/modbus"
	"net"
	"sync"
	"time"
)

// ProxyInstance represents a running proxy.
type ProxyInstance struct {
	ID         string
	Name       string
	ListenAddr string
	TargetAddr string

	listener      net.Listener
	targetConn    net.Conn
	targetMu      sync.Mutex

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
func NewProxyInstance(id, name, listen, target string, l *logger.Logger, tracker *devices.Tracker) *ProxyInstance {
	return &ProxyInstance{
		ID:            id,
		Name:          name,
		ListenAddr:    listen,
		TargetAddr:    target,
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
		p.listener.Close()
	}
	p.closeTargetConn()
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
	p.targetMu.Lock()
	defer p.targetMu.Unlock()

	if err := p.ensureTargetConn(); err != nil {
		return nil, err
	}

	_ = p.targetConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := p.targetConn.Write(req); err != nil {
		p.closeTargetConn()
		if err := p.ensureTargetConn(); err != nil {
			return nil, err
		}
		_ = p.targetConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err := p.targetConn.Write(req); err != nil {
			p.closeTargetConn()
			return nil, err
		}
	}

	_ = p.targetConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp, err := modbus.ReadFrame(p.targetConn)
	if err != nil {
		p.closeTargetConn()
		return nil, err
	}
	return resp, nil
}

func (p *ProxyInstance) ensureTargetConn() error {
	if p.targetConn != nil {
		return nil
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(p.ctx, "tcp", p.TargetAddr)
	if err != nil {
		return err
	}
	p.targetConn = conn
	return nil
}

func (p *ProxyInstance) closeTargetConn() {
	if p.targetConn != nil {
		p.targetConn.Close()
		p.targetConn = nil
	}
}
