package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Collector collects system-wide metrics for monitoring and observability.
type Collector struct {
	startTime time.Time

	// Request metrics
	totalRequests uint64
	totalErrors   uint64
	totalBytes    uint64

	// Connection metrics
	activeConnections  int64
	totalConnections   uint64
	rejectedConnections uint64

	// Latency tracking (in microseconds)
	mu               sync.RWMutex
	latencies        []uint64 // Ring buffer for recent latencies
	latencyIndex     int
	latencyCount     int
	latencyBufferSize int

	// Per-proxy metrics
	proxyMetrics sync.Map // map[string]*ProxyMetrics
}

// ProxyMetrics holds metrics for a single proxy.
type ProxyMetrics struct {
	Name              string
	Requests          uint64
	Errors            uint64
	BytesReceived     uint64
	BytesSent         uint64
	ActiveConnections int64
	TotalConnections  uint64
	AvgLatencyUs      uint64 // Average latency in microseconds

	// Latency histogram (microseconds)
	mu         sync.RWMutex
	latencies  []uint64
	latencyIdx int
	latencyLen int
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		startTime:         time.Now(),
		latencyBufferSize: 1000, // Keep last 1000 latencies
		latencies:         make([]uint64, 1000),
	}
}

// RecordRequest records a successful request.
func (c *Collector) RecordRequest(latencyUs uint64, bytes uint64) {
	atomic.AddUint64(&c.totalRequests, 1)
	atomic.AddUint64(&c.totalBytes, bytes)

	c.mu.Lock()
	c.latencies[c.latencyIndex] = latencyUs
	c.latencyIndex = (c.latencyIndex + 1) % c.latencyBufferSize
	if c.latencyCount < c.latencyBufferSize {
		c.latencyCount++
	}
	c.mu.Unlock()
}

// RecordError records a failed request.
func (c *Collector) RecordError() {
	atomic.AddUint64(&c.totalErrors, 1)
}

// RecordConnection records a new connection.
func (c *Collector) RecordConnection() {
	atomic.AddInt64(&c.activeConnections, 1)
	atomic.AddUint64(&c.totalConnections, 1)
}

// RecordDisconnection records a connection being closed.
func (c *Collector) RecordDisconnection() {
	atomic.AddInt64(&c.activeConnections, -1)
}

// RecordRejectedConnection records a rejected connection.
func (c *Collector) RecordRejectedConnection() {
	atomic.AddUint64(&c.rejectedConnections, 1)
}

// GetProxyMetrics returns metrics for a specific proxy, creating if needed.
func (c *Collector) GetProxyMetrics(proxyID, proxyName string) *ProxyMetrics {
	if m, ok := c.proxyMetrics.Load(proxyID); ok {
		return m.(*ProxyMetrics)
	}

	// Create new proxy metrics
	pm := &ProxyMetrics{
		Name:      proxyName,
		latencies: make([]uint64, 100), // Keep last 100 latencies per proxy
	}

	c.proxyMetrics.Store(proxyID, pm)
	return pm
}

// Snapshot returns a snapshot of current metrics.
func (c *Collector) Snapshot() *Snapshot {
	snapshot := &Snapshot{
		Timestamp:           time.Now(),
		Uptime:              time.Since(c.startTime),
		TotalRequests:       atomic.LoadUint64(&c.totalRequests),
		TotalErrors:         atomic.LoadUint64(&c.totalErrors),
		TotalBytes:          atomic.LoadUint64(&c.totalBytes),
		ActiveConnections:   atomic.LoadInt64(&c.activeConnections),
		TotalConnections:    atomic.LoadUint64(&c.totalConnections),
		RejectedConnections: atomic.LoadUint64(&c.rejectedConnections),
	}

	// Calculate error rate
	if snapshot.TotalRequests > 0 {
		snapshot.ErrorRate = float64(snapshot.TotalErrors) / float64(snapshot.TotalRequests)
	}

	// Calculate requests per second
	if snapshot.Uptime > 0 {
		snapshot.RequestsPerSec = float64(snapshot.TotalRequests) / snapshot.Uptime.Seconds()
	}

	// Calculate latency percentiles
	c.mu.RLock()
	if c.latencyCount > 0 {
		latencies := make([]uint64, c.latencyCount)
		copy(latencies, c.latencies[:c.latencyCount])
		c.mu.RUnlock()

		snapshot.LatencyStats = calculateLatencyStats(latencies)
	} else {
		c.mu.RUnlock()
	}

	// Get runtime stats
	snapshot.RuntimeStats = getRuntimeStats()

	// Get per-proxy metrics
	snapshot.ProxyMetrics = make(map[string]ProxyMetricSnapshot)
	c.proxyMetrics.Range(func(key, value interface{}) bool {
		proxyID := key.(string)
		pm := value.(*ProxyMetrics)

		snapshot.ProxyMetrics[proxyID] = ProxyMetricSnapshot{
			Name:              pm.Name,
			Requests:          atomic.LoadUint64(&pm.Requests),
			Errors:            atomic.LoadUint64(&pm.Errors),
			BytesReceived:     atomic.LoadUint64(&pm.BytesReceived),
			BytesSent:         atomic.LoadUint64(&pm.BytesSent),
			ActiveConnections: atomic.LoadInt64(&pm.ActiveConnections),
			TotalConnections:  atomic.LoadUint64(&pm.TotalConnections),
			AvgLatencyUs:      atomic.LoadUint64(&pm.AvgLatencyUs),
		}

		return true
	})

	return snapshot
}

// RecordProxyRequest records a request for a specific proxy.
func (pm *ProxyMetrics) RecordRequest(latencyUs uint64, bytesReceived, bytesSent uint64) {
	atomic.AddUint64(&pm.Requests, 1)
	atomic.AddUint64(&pm.BytesReceived, bytesReceived)
	atomic.AddUint64(&pm.BytesSent, bytesSent)

	pm.mu.Lock()
	pm.latencies[pm.latencyIdx] = latencyUs
	pm.latencyIdx = (pm.latencyIdx + 1) % len(pm.latencies)
	if pm.latencyLen < len(pm.latencies) {
		pm.latencyLen++
	}
	pm.mu.Unlock()

	// Update average latency
	pm.updateAvgLatency()
}

// RecordProxyError records an error for a specific proxy.
func (pm *ProxyMetrics) RecordProxyError() {
	atomic.AddUint64(&pm.Errors, 1)
}

// RecordProxyConnection records a new connection for a proxy.
func (pm *ProxyMetrics) RecordProxyConnection() {
	atomic.AddInt64(&pm.ActiveConnections, 1)
	atomic.AddUint64(&pm.TotalConnections, 1)
}

// RecordProxyDisconnection records a disconnection for a proxy.
func (pm *ProxyMetrics) RecordProxyDisconnection() {
	atomic.AddInt64(&pm.ActiveConnections, -1)
}

// updateAvgLatency calculates and updates the average latency.
func (pm *ProxyMetrics) updateAvgLatency() {
	pm.mu.RLock()
	if pm.latencyLen == 0 {
		pm.mu.RUnlock()
		return
	}

	var sum uint64
	for i := 0; i < pm.latencyLen; i++ {
		sum += pm.latencies[i]
	}
	avg := sum / uint64(pm.latencyLen)
	pm.mu.RUnlock()

	atomic.StoreUint64(&pm.AvgLatencyUs, avg)
}

// Snapshot holds a point-in-time snapshot of metrics.
type Snapshot struct {
	Timestamp           time.Time
	Uptime              time.Duration
	TotalRequests       uint64
	TotalErrors         uint64
	TotalBytes          uint64
	ActiveConnections   int64
	TotalConnections    uint64
	RejectedConnections uint64
	ErrorRate           float64
	RequestsPerSec      float64
	LatencyStats        LatencyStats
	RuntimeStats        RuntimeStats
	ProxyMetrics        map[string]ProxyMetricSnapshot
}

// ProxyMetricSnapshot is a snapshot of proxy metrics.
type ProxyMetricSnapshot struct {
	Name              string
	Requests          uint64
	Errors            uint64
	BytesReceived     uint64
	BytesSent         uint64
	ActiveConnections int64
	TotalConnections  uint64
	AvgLatencyUs      uint64
}

// LatencyStats holds latency statistics.
type LatencyStats struct {
	Min uint64
	Max uint64
	Avg uint64
	P50 uint64
	P95 uint64
	P99 uint64
}

// RuntimeStats holds Go runtime statistics.
type RuntimeStats struct {
	Goroutines   int
	HeapAlloc    uint64
	HeapInuse    uint64
	HeapObjects  uint64
	GCRuns       uint32
	LastGCPause  time.Duration
	MemoryAlloc  uint64
	MemoryTotal  uint64
	MemorySys    uint64
}

// getRuntimeStats collects Go runtime statistics.
func getRuntimeStats() RuntimeStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return RuntimeStats{
		Goroutines:   runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapInuse:    m.HeapInuse,
		HeapObjects:  m.HeapObjects,
		GCRuns:       m.NumGC,
		LastGCPause:  time.Duration(m.PauseNs[(m.NumGC+255)%256]),
		MemoryAlloc:  m.Alloc,
		MemoryTotal:  m.TotalAlloc,
		MemorySys:    m.Sys,
	}
}

// calculateLatencyStats calculates latency percentiles.
// Note: This modifies the input slice by sorting it.
func calculateLatencyStats(latencies []uint64) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{}
	}

	// Simple bubble sort (good enough for small samples)
	// For production, use sort.Slice from standard library
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	stats := LatencyStats{
		Min: latencies[0],
		Max: latencies[len(latencies)-1],
	}

	// Calculate average
	var sum uint64
	for _, l := range latencies {
		sum += l
	}
	stats.Avg = sum / uint64(len(latencies))

	// Calculate percentiles
	stats.P50 = latencies[len(latencies)*50/100]
	stats.P95 = latencies[len(latencies)*95/100]
	stats.P99 = latencies[len(latencies)*99/100]

	return stats
}

// GlobalCollector is the global metrics collector instance.
var GlobalCollector = NewCollector()
