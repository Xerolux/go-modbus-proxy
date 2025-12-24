package performance

import (
	"context"
	"net"
	"sync"
	"time"
)

// DNSCache provides a simple in-memory DNS cache to reduce lookup latency.
// This is particularly useful when connecting to the same targets repeatedly.
type DNSCache struct {
	mu      sync.RWMutex
	cache   map[string]*dnsEntry
	ttl     time.Duration
	maxSize int

	// Stats
	hits   uint64
	misses uint64
}

type dnsEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

// NewDNSCache creates a new DNS cache with the specified TTL and max size.
func NewDNSCache(ttl time.Duration, maxSize int) *DNSCache {
	cache := &DNSCache{
		cache:   make(map[string]*dnsEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// DefaultDNSCache creates a DNS cache with reasonable defaults.
func DefaultDNSCache() *DNSCache {
	return NewDNSCache(5*time.Minute, 1000)
}

// LookupHost performs a DNS lookup with caching.
func (c *DNSCache) LookupHost(ctx context.Context, host string) ([]net.IP, error) {
	// Check cache first
	c.mu.RLock()
	entry, found := c.cache[host]
	c.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		return entry.ips, nil
	}

	// Cache miss - perform actual lookup
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict old entries if cache is full
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[host] = &dnsEntry{
		ips:       ips,
		expiresAt: time.Now().Add(c.ttl),
	}

	return ips, nil
}

// Dial performs a dial with DNS caching.
func (c *DNSCache) Dial(network, address string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, address)
}

// DialContext performs a dial with DNS caching and context.
func (c *DNSCache) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	// If it's already an IP, no need to cache
	if ip := net.ParseIP(host); ip != nil {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, network, address)
	}

	// Lookup with caching
	ips, err := c.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, &net.DNSError{
			Err:  "no IP addresses found",
			Name: host,
		}
	}

	// Try to connect to the first IP
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
}

// Clear removes all entries from the cache.
func (c *DNSCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*dnsEntry)
}

// Stats returns cache statistics.
func (c *DNSCache) Stats() (hits, misses uint64, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, len(c.cache)
}

// HitRate returns the cache hit rate as a percentage.
func (c *DNSCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0
	}

	return float64(c.hits) / float64(total) * 100
}

// cleanup periodically removes expired entries.
func (c *DNSCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for host, entry := range c.cache {
			if now.After(entry.expiresAt) {
				delete(c.cache, host)
			}
		}
		c.mu.Unlock()
	}
}

// evictOldest removes the oldest entry from the cache.
// Must be called with lock held.
func (c *DNSCache) evictOldest() {
	var oldestHost string
	var oldestTime time.Time

	first := true
	for host, entry := range c.cache {
		if first || entry.expiresAt.Before(oldestTime) {
			oldestHost = host
			oldestTime = entry.expiresAt
			first = false
		}
	}

	if oldestHost != "" {
		delete(c.cache, oldestHost)
	}
}
