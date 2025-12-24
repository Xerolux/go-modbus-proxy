package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"modbusproxy/pkg/performance"
)

// OptimizedProxy is a high-performance Modbus TCP proxy with all optimizations enabled.
// Target: 10,000+ concurrent connections, 1M+ requests/sec, <100MB memory, <1ms latency
type OptimizedProxy struct {
	id     string
	name   string
	listen string
	target string

	// Performance components
	dnsCache   *performance.DNSCache
	bufferPool *performance.BufferPool
	keepAlive  performance.KeepAliveConfig
	tcpConfig  performance.OptimizedTCPConfig

	// Connection management
	listener      net.Listener
	activeConns   int64
	totalConns    uint64
	totalRequests uint64
	totalErrors   uint64

	// Control
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    atomic.Bool
	maxConns   int64
	connSem    chan struct{} // Semaphore for connection limiting

	// Metrics
	startTime time.Time
}

// OptimizedConfig holds configuration for the optimized proxy.
type OptimizedConfig struct {
	ID         string
	Name       string
	ListenAddr string
	TargetAddr string
	MaxConns   int64 // Maximum concurrent connections (0 = unlimited)

	// Performance settings
	EnableDNSCache    bool
	EnableBufferPool  bool
	EnableKeepAlive   bool
	EnableTCPOptimize bool
}

// NewOptimizedProxy creates a new high-performance proxy with all optimizations.
func NewOptimizedProxy(config OptimizedConfig) *OptimizedProxy {
	if config.MaxConns <= 0 {
		config.MaxConns = 10000 // Default: 10k concurrent connections
	}

	p := &OptimizedProxy{
		id:         config.ID,
		name:       config.Name,
		listen:     config.ListenAddr,
		target:     config.TargetAddr,
		maxConns:   config.MaxConns,
		connSem:    make(chan struct{}, config.MaxConns),
		startTime:  time.Now(),
	}

	// Initialize performance components
	if config.EnableDNSCache {
		p.dnsCache = performance.DefaultDNSCache()
	}

	if config.EnableBufferPool {
		p.bufferPool = performance.NewBufferPool(512) // 512B buffers for Modbus
	}

	if config.EnableKeepAlive {
		p.keepAlive = performance.DefaultKeepAliveConfig()
	}

	if config.EnableTCPOptimize {
		p.tcpConfig = performance.DefaultTCPConfig()
	}

	return p
}

// Start starts the proxy server.
func (p *OptimizedProxy) Start(ctx context.Context) error {
	if !p.running.CompareAndSwap(false, true) {
		return fmt.Errorf("proxy already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)

	// Create listener
	listener, err := net.Listen("tcp", p.listen)
	if err != nil {
		p.running.Store(false)
		return fmt.Errorf("failed to listen on %s: %w", p.listen, err)
	}

	p.listener = listener
	log.Printf("[%s] Optimized proxy started on %s -> %s (max conns: %d)",
		p.id, p.listen, p.target, p.maxConns)

	// Accept connections
	p.wg.Add(1)
	go p.acceptLoop()

	return nil
}

// Stop gracefully stops the proxy.
func (p *OptimizedProxy) Stop() error {
	if !p.running.CompareAndSwap(true, false) {
		return fmt.Errorf("proxy not running")
	}

	log.Printf("[%s] Stopping proxy...", p.id)

	// Stop accepting new connections
	if p.listener != nil {
		p.listener.Close()
	}

	// Cancel context to stop all goroutines
	if p.cancel != nil {
		p.cancel()
	}

	// Wait for all connections to finish
	p.wg.Wait()

	log.Printf("[%s] Proxy stopped", p.id)
	return nil
}

// acceptLoop accepts incoming connections.
func (p *OptimizedProxy) acceptLoop() {
	defer p.wg.Done()

	for {
		conn, err := p.listener.Accept()
		if err != nil {
			if !p.running.Load() {
				return // Shutting down
			}
			log.Printf("[%s] Accept error: %v", p.id, err)
			continue
		}

		// Try to acquire connection slot
		select {
		case p.connSem <- struct{}{}:
			// Got slot, handle connection
			p.wg.Add(1)
			go p.handleConnection(conn)
		default:
			// No slots available, reject connection
			conn.Close()
			atomic.AddUint64(&p.totalErrors, 1)
			log.Printf("[%s] Connection rejected (max connections reached)", p.id)
		}
	}
}

// handleConnection handles a single client connection.
func (p *OptimizedProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer func() {
		clientConn.Close()
		<-p.connSem // Release connection slot
		atomic.AddInt64(&p.activeConns, -1)
	}()

	atomic.AddInt64(&p.activeConns, 1)
	atomic.AddUint64(&p.totalConns, 1)

	// Apply TCP optimizations
	if p.tcpConfig.NoDelay {
		performance.ConfigureTCPConnection(clientConn, p.tcpConfig)
	}

	// Apply keep-alive
	if p.keepAlive.Enabled {
		performance.ConfigureKeepAlive(clientConn, p.keepAlive)
	}

	// Connect to target
	var targetConn net.Conn
	var err error

	if p.dnsCache != nil {
		// Use DNS cache
		targetConn, err = p.dnsCache.DialContext(p.ctx, "tcp", p.target)
	} else {
		// Direct dial
		dialer := net.Dialer{Timeout: 5 * time.Second}
		targetConn, err = dialer.DialContext(p.ctx, "tcp", p.target)
	}

	if err != nil {
		atomic.AddUint64(&p.totalErrors, 1)
		return
	}
	defer targetConn.Close()

	// Apply optimizations to target connection
	if p.tcpConfig.NoDelay {
		performance.ConfigureTCPConnection(targetConn, p.tcpConfig)
	}
	if p.keepAlive.Enabled {
		performance.ConfigureKeepAlive(targetConn, p.keepAlive)
	}

	// Bidirectional copy with buffer pooling
	p.bidirectionalCopy(clientConn, targetConn)
}

// bidirectionalCopy performs optimized bidirectional data copying.
func (p *OptimizedProxy) bidirectionalCopy(client, target net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Target
	go func() {
		defer wg.Done()
		p.copyData(target, client, "client->target")
	}()

	// Target -> Client
	go func() {
		defer wg.Done()
		p.copyData(client, target, "target->client")
	}()

	wg.Wait()
}

// copyData copies data from src to dst using pooled buffers.
func (p *OptimizedProxy) copyData(dst, src net.Conn, direction string) {
	var buf []byte
	if p.bufferPool != nil {
		buf = p.bufferPool.Get()
		defer p.bufferPool.Put(buf)
	} else {
		buf = make([]byte, 512)
	}

	for {
		// Read from source
		n, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				atomic.AddUint64(&p.totalErrors, 1)
			}
			return
		}

		atomic.AddUint64(&p.totalRequests, 1)

		// Write to destination
		_, err = dst.Write(buf[:n])
		if err != nil {
			atomic.AddUint64(&p.totalErrors, 1)
			return
		}
	}
}

// Stats returns current proxy statistics.
func (p *OptimizedProxy) Stats() ProxyStats {
	uptime := time.Since(p.startTime)
	active := atomic.LoadInt64(&p.activeConns)
	total := atomic.LoadUint64(&p.totalConns)
	requests := atomic.LoadUint64(&p.totalRequests)
	errors := atomic.LoadUint64(&p.totalErrors)

	var reqPerSec float64
	if uptime.Seconds() > 0 {
		reqPerSec = float64(requests) / uptime.Seconds()
	}

	var errorRate float64
	if requests > 0 {
		errorRate = float64(errors) / float64(requests)
	}

	return ProxyStats{
		ID:                p.id,
		Name:              p.name,
		Uptime:            uptime,
		ActiveConnections: active,
		TotalConnections:  total,
		TotalRequests:     requests,
		TotalErrors:       errors,
		RequestsPerSec:    reqPerSec,
		ErrorRate:         errorRate,
		Running:           p.running.Load(),
	}
}

// ProxyStats holds proxy statistics.
type ProxyStats struct {
	ID                string
	Name              string
	Uptime            time.Duration
	ActiveConnections int64
	TotalConnections  uint64
	TotalRequests     uint64
	TotalErrors       uint64
	RequestsPerSec    float64
	ErrorRate         float64
	Running           bool
}

// String returns a human-readable representation of stats.
func (s ProxyStats) String() string {
	return fmt.Sprintf(
		"[%s] Active: %d, Total: %d, Requests: %d, Errors: %d, RPS: %.2f, Error Rate: %.2f%%, Uptime: %s",
		s.ID, s.ActiveConnections, s.TotalConnections, s.TotalRequests,
		s.TotalErrors, s.RequestsPerSec, s.ErrorRate*100, s.Uptime,
	)
}
