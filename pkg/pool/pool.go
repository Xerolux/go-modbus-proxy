package pool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	// ErrPoolClosed is returned when the pool is closed.
	ErrPoolClosed = errors.New("connection pool is closed")
	// ErrPoolExhausted is returned when the pool is exhausted.
	ErrPoolExhausted = errors.New("connection pool exhausted")
)

// Config holds pool configuration.
type Config struct {
	// Initial size of the pool
	InitialSize int
	// Maximum size of the pool
	MaxSize int
	// Maximum idle time before connection is closed
	MaxIdleTime time.Duration
	// Timeout for acquiring a connection
	AcquireTimeout time.Duration
	// Dialer function
	Dialer func(context.Context) (net.Conn, error)
}

// Pool is a connection pool.
type Pool struct {
	mu      sync.Mutex
	conns   chan *poolConn
	factory func(context.Context) (net.Conn, error)

	maxIdleTime    time.Duration
	acquireTimeout time.Duration

	closed bool
	size   int
	maxSize int
}

type poolConn struct {
	conn       net.Conn
	lastUsed   time.Time
	inUse      bool
}

// NewPool creates a new connection pool.
func NewPool(cfg Config) (*Pool, error) {
	if cfg.InitialSize < 0 {
		cfg.InitialSize = 0
	}
	if cfg.MaxSize < 1 {
		cfg.MaxSize = 10
	}
	if cfg.MaxIdleTime == 0 {
		cfg.MaxIdleTime = 5 * time.Minute
	}
	if cfg.AcquireTimeout == 0 {
		cfg.AcquireTimeout = 30 * time.Second
	}

	p := &Pool{
		conns:          make(chan *poolConn, cfg.MaxSize),
		factory:        cfg.Dialer,
		maxIdleTime:    cfg.MaxIdleTime,
		acquireTimeout: cfg.AcquireTimeout,
		maxSize:        cfg.MaxSize,
	}

	// Pre-populate pool
	for i := 0; i < cfg.InitialSize; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		conn, err := p.factory(ctx)
		cancel()

		if err != nil {
			return nil, err
		}

		p.conns <- &poolConn{
			conn:     conn,
			lastUsed: time.Now(),
		}
		p.size++
	}

	// Start cleanup goroutine
	go p.cleanup()

	return p, nil
}

// Get acquires a connection from the pool.
func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}
	p.mu.Unlock()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.acquireTimeout)
		defer cancel()
		deadline, _ = ctx.Deadline()
	}

	timeout := time.Until(deadline)
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, ErrPoolExhausted
		case pc := <-p.conns:
			// Check if connection is still valid
			if time.Since(pc.lastUsed) > p.maxIdleTime {
				pc.conn.Close()
				p.mu.Lock()
				p.size--
				p.mu.Unlock()
				continue
			}

			pc.inUse = true
			return &wrappedConn{
				Conn: pc.conn,
				pool: p,
				pc:   pc,
			}, nil

		default:
			// Try to create new connection
			p.mu.Lock()
			if p.size < p.maxSize {
				p.size++
				p.mu.Unlock()

				conn, err := p.factory(ctx)
				if err != nil {
					p.mu.Lock()
					p.size--
					p.mu.Unlock()
					return nil, err
				}

				pc := &poolConn{
					conn:     conn,
					lastUsed: time.Now(),
					inUse:    true,
				}

				return &wrappedConn{
					Conn: conn,
					pool: p,
					pc:   pc,
				}, nil
			}
			p.mu.Unlock()

			// Wait a bit before retrying
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// put returns a connection to the pool.
func (p *Pool) put(pc *poolConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		pc.conn.Close()
		p.size--
		return
	}

	pc.lastUsed = time.Now()
	pc.inUse = false

	select {
	case p.conns <- pc:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		pc.conn.Close()
		p.size--
	}
}

// Close closes the pool and all connections.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.conns)

	for pc := range p.conns {
		pc.conn.Close()
	}

	return nil
}

// Stats returns pool statistics.
func (p *Pool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	return PoolStats{
		TotalConns:     p.size,
		IdleConns:      len(p.conns),
		ActiveConns:    p.size - len(p.conns),
		MaxSize:        p.maxSize,
	}
}

// PoolStats represents pool statistics.
type PoolStats struct {
	TotalConns  int
	IdleConns   int
	ActiveConns int
	MaxSize     int
}

// cleanup periodically cleans up idle connections.
func (p *Pool) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return
		}
		p.mu.Unlock()

		// Remove expired connections
		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case <-timeout:
				return
			case pc := <-p.conns:
				if time.Since(pc.lastUsed) > p.maxIdleTime {
					pc.conn.Close()
					p.mu.Lock()
					p.size--
					p.mu.Unlock()
				} else {
					// Put it back
					p.conns <- pc
					return
				}
			default:
				return
			}
		}
	}
}

// wrappedConn wraps a connection to return it to the pool on close.
type wrappedConn struct {
	net.Conn
	pool *Pool
	pc   *poolConn
	once sync.Once
}

func (w *wrappedConn) Close() error {
	w.once.Do(func() {
		w.pool.put(w.pc)
	})
	return nil
}
