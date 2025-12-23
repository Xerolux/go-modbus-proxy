package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks application metrics.
type Metrics struct {
	mu sync.RWMutex

	// Counters (atomic)
	totalRequests   atomic.Int64
	totalErrors     atomic.Int64
	totalConnections atomic.Int64

	// Per-proxy metrics
	proxyMetrics map[string]*ProxyMetrics

	// System metrics
	startTime time.Time
}

// ProxyMetrics tracks metrics for a single proxy.
type ProxyMetrics struct {
	requests   atomic.Int64
	errors     atomic.Int64
	bytesIn    atomic.Int64
	bytesOut   atomic.Int64
	activeConns atomic.Int32

	// Latency tracking
	mu              sync.Mutex
	latencies       []time.Duration
	maxLatencyCount int
}

// NewMetrics creates a new metrics tracker.
func NewMetrics() *Metrics {
	return &Metrics{
		proxyMetrics: make(map[string]*ProxyMetrics),
		startTime:    time.Now(),
	}
}

// RegisterProxy registers a new proxy for metrics tracking.
func (m *Metrics) RegisterProxy(proxyID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.proxyMetrics[proxyID]; !exists {
		m.proxyMetrics[proxyID] = &ProxyMetrics{
			maxLatencyCount: 1000, // Keep last 1000 latencies
			latencies:       make([]time.Duration, 0, 1000),
		}
	}
}

// UnregisterProxy removes a proxy from metrics tracking.
func (m *Metrics) UnregisterProxy(proxyID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.proxyMetrics, proxyID)
}

// RecordRequest records a successful request.
func (m *Metrics) RecordRequest(proxyID string, latency time.Duration, bytesIn, bytesOut int64) {
	m.totalRequests.Add(1)

	m.mu.RLock()
	pm, exists := m.proxyMetrics[proxyID]
	m.mu.RUnlock()

	if exists {
		pm.requests.Add(1)
		pm.bytesIn.Add(bytesIn)
		pm.bytesOut.Add(bytesOut)
		pm.recordLatency(latency)
	}
}

// RecordError records an error.
func (m *Metrics) RecordError(proxyID string) {
	m.totalErrors.Add(1)

	m.mu.RLock()
	pm, exists := m.proxyMetrics[proxyID]
	m.mu.RUnlock()

	if exists {
		pm.errors.Add(1)
	}
}

// RecordConnection records a new connection.
func (m *Metrics) RecordConnection(proxyID string, active bool) {
	if active {
		m.totalConnections.Add(1)
	}

	m.mu.RLock()
	pm, exists := m.proxyMetrics[proxyID]
	m.mu.RUnlock()

	if exists {
		if active {
			pm.activeConns.Add(1)
		} else {
			pm.activeConns.Add(-1)
		}
	}
}

// recordLatency records a latency measurement.
func (pm *ProxyMetrics) recordLatency(latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.latencies) >= pm.maxLatencyCount {
		// Remove oldest
		pm.latencies = pm.latencies[1:]
	}
	pm.latencies = append(pm.latencies, latency)
}

// GetStats returns current statistics.
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		TotalRequests:    m.totalRequests.Load(),
		TotalErrors:      m.totalErrors.Load(),
		TotalConnections: m.totalConnections.Load(),
		Uptime:           time.Since(m.startTime),
		ProxyStats:       make(map[string]ProxyStats),
	}

	for proxyID, pm := range m.proxyMetrics {
		stats.ProxyStats[proxyID] = pm.getStats()
	}

	return stats
}

// getStats returns statistics for this proxy.
func (pm *ProxyMetrics) getStats() ProxyStats {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stats := ProxyStats{
		Requests:    pm.requests.Load(),
		Errors:      pm.errors.Load(),
		BytesIn:     pm.bytesIn.Load(),
		BytesOut:    pm.bytesOut.Load(),
		ActiveConns: int(pm.activeConns.Load()),
	}

	// Calculate latency percentiles
	if len(pm.latencies) > 0 {
		sorted := make([]time.Duration, len(pm.latencies))
		copy(sorted, pm.latencies)

		// Simple percentile calculation
		stats.LatencyP50 = sorted[len(sorted)*50/100]
		stats.LatencyP95 = sorted[len(sorted)*95/100]
		stats.LatencyP99 = sorted[len(sorted)*99/100]

		// Calculate average
		var sum time.Duration
		for _, l := range sorted {
			sum += l
		}
		stats.LatencyAvg = sum / time.Duration(len(sorted))
	}

	return stats
}

// Stats represents overall statistics.
type Stats struct {
	TotalRequests    int64
	TotalErrors      int64
	TotalConnections int64
	Uptime           time.Duration
	ProxyStats       map[string]ProxyStats
}

// ProxyStats represents statistics for a single proxy.
type ProxyStats struct {
	Requests    int64
	Errors      int64
	BytesIn     int64
	BytesOut    int64
	ActiveConns int
	LatencyAvg  time.Duration
	LatencyP50  time.Duration
	LatencyP95  time.Duration
	LatencyP99  time.Duration
}
