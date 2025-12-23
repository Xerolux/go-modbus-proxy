package proxy

import (
	"log"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/modbus"
	"net"
	"sync"
	"time"
)

// Proxy implements the Modbus TCP proxy.
type Proxy struct {
	cfg        config.Config
	listener   net.Listener
	targetConn net.Conn
	targetMu   sync.Mutex
	stopChan   chan struct{}
	running    bool
}

// NewProxy creates a new proxy instance.
func NewProxy(cfg config.Config) *Proxy {
	return &Proxy{
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

// Start starts the proxy listener. It blocks until Stop is called or an error occurs.
func (p *Proxy) Start() error {
	l, err := net.Listen("tcp", p.cfg.ListenAddr)
	if err != nil {
		return err
	}
	p.listener = l
	p.running = true
	log.Printf("Proxy listening on %s, target: %s", p.cfg.ListenAddr, p.cfg.TargetAddr)

	go func() {
		<-p.stopChan
		p.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if !p.running {
				return nil // Expected shutdown
			}
			log.Printf("Accept error: %v", err)
			continue
		}
		go p.handleClient(conn)
	}
}

// Stop stops the proxy.
func (p *Proxy) Stop() {
	if p.running {
		p.running = false
		close(p.stopChan)
	}
}

// Close closes the listener and target connection.
func (p *Proxy) Close() {
	if p.listener != nil {
		p.listener.Close()
	}
	p.targetMu.Lock()
	if p.targetConn != nil {
		p.targetConn.Close()
		p.targetConn = nil
	}
	p.targetMu.Unlock()
}

func (p *Proxy) handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	for {
		// Read request from client
		reqFrame, err := modbus.ReadFrame(clientConn)
		if err != nil {
			// Start of EOF is expected if client closes connection
			return
		}

		// Process request
		respFrame, err := p.forwardRequest(reqFrame)
		if err != nil {
			log.Printf("Forward error: %v", err)
			return
		}

		// Send response to client
		if _, err := clientConn.Write(respFrame); err != nil {
			log.Printf("Write response error: %v", err)
			return
		}
	}
}

func (p *Proxy) forwardRequest(req []byte) ([]byte, error) {
	p.targetMu.Lock()
	defer p.targetMu.Unlock()

	// Ensure target connection
	if err := p.ensureTargetConn(); err != nil {
		return nil, err
	}

	// Write request
	p.targetConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := p.targetConn.Write(req); err != nil {
		p.closeTargetConn()
		// Try to reconnect once
		if err := p.ensureTargetConn(); err != nil {
			return nil, err
		}
		// Retry write
		p.targetConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err := p.targetConn.Write(req); err != nil {
			p.closeTargetConn()
			return nil, err
		}
	}

	// Read response
	p.targetConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	resp, err := modbus.ReadFrame(p.targetConn)
	if err != nil {
		p.closeTargetConn()
		return nil, err
	}

	return resp, nil
}

func (p *Proxy) ensureTargetConn() error {
	if p.targetConn != nil {
		return nil
	}

	conn, err := net.DialTimeout("tcp", p.cfg.TargetAddr, 5*time.Second)
	if err != nil {
		return err
	}
	p.targetConn = conn
	return nil
}

func (p *Proxy) closeTargetConn() {
	if p.targetConn != nil {
		p.targetConn.Close()
		p.targetConn = nil
	}
}
