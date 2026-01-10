package proxy

import (
	"context"
	"fmt"
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

	// Config
	MaxReadSize int

	listener   net.Listener
	targetConn net.Conn
	targetMu   sync.Mutex

	log           *logger.Logger
	deviceTracker *devices.Tracker
	ctx           context.Context
	cancel        context.CancelFunc

	Stats Stats
}

type Stats struct {
	Uptime    time.Duration
	LastStart time.Time
	Requests  int64
	Errors    int64
	Status    string // "Running", "Stopped", "Error"
}

// NewProxyInstance creates a new proxy.
func NewProxyInstance(id, name, listen, target string, maxReadSize int, l *logger.Logger, tracker *devices.Tracker) *ProxyInstance {
	return &ProxyInstance{
		ID:            id,
		Name:          name,
		ListenAddr:    listen,
		TargetAddr:    target,
		MaxReadSize:   maxReadSize,
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

		var respFrame []byte
		var errFwd error

		// Check if splitting is needed
		if p.MaxReadSize > 0 && modbus.IsReadRequest(reqFrame) {
			respFrame, errFwd = p.handleSplitRead(reqFrame)
		} else {
			respFrame, errFwd = p.forwardRequest(reqFrame)
		}

		if errFwd != nil {
			p.log.Error(p.ID, "Forward error: "+errFwd.Error())
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
	var aggregatedData []byte
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

func (p *ProxyInstance) forwardRequest(req []byte) ([]byte, error) {
	p.targetMu.Lock()
	defer p.targetMu.Unlock()
	return p.forwardRequestLocked(req)
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
