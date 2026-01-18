package metrics

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks application metrics.
type Metrics struct {
	mu sync.RWMutex

	// Counters (atomic)
	totalRequests    atomic.Int64
	totalErrors      atomic.Int64
	totalConnections atomic.Int64

	// Per-proxy metrics
	proxyMetrics map[string]*ProxyMetrics

	// System metrics
	startTime time.Time
}

// ProxyMetrics tracks metrics for a single proxy.
type ProxyMetrics struct {
	requests    atomic.Int64
	errors      atomic.Int64
	bytesIn     atomic.Int64
	bytesOut    atomic.Int64
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

// GetPrometheusMetrics returns metrics in Prometheus format.
func (m *Metrics) GetPrometheusMetrics() string {
	stats := m.GetStats()

	var output strings.Builder

	// System-wide metrics
	output.WriteString(fmt.Sprintf("# HELP modbridge_total_requests Total number of requests\n"))
	output.WriteString(fmt.Sprintf("# TYPE modbridge_total_requests counter\n"))
	output.WriteString(fmt.Sprintf("modbridge_total_requests %d\n\n", stats.TotalRequests))

	output.WriteString(fmt.Sprintf("# HELP modbridge_total_errors Total number of errors\n"))
	output.WriteString(fmt.Sprintf("# TYPE modbridge_total_errors counter\n"))
	output.WriteString(fmt.Sprintf("modbridge_total_errors %d\n\n", stats.TotalErrors))

	output.WriteString(fmt.Sprintf("# HELP modbridge_total_connections Total number of connections\n"))
	output.WriteString(fmt.Sprintf("# TYPE modbridge_total_connections counter\n"))
	output.WriteString(fmt.Sprintf("modbridge_total_connections %d\n\n", stats.TotalConnections))

	output.WriteString(fmt.Sprintf("# HELP modbridge_uptime_seconds Uptime in seconds\n"))
	output.WriteString(fmt.Sprintf("# TYPE modbridge_uptime_seconds gauge\n"))
	output.WriteString(fmt.Sprintf("modbridge_uptime_seconds %f\n\n", stats.Uptime.Seconds()))

	// Per-proxy metrics
	for proxyID, proxyStats := range stats.ProxyStats {
		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_requests_total Total requests for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_requests_total counter\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_requests_total{proxy_id="%s"} %d\n\n`, proxyID, proxyStats.Requests))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_errors_total Total errors for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_errors_total counter\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_errors_total{proxy_id="%s"} %d\n\n`, proxyID, proxyStats.Errors))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_active_connections Active connections for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_active_connections gauge\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_active_connections{proxy_id="%s"} %d\n\n`, proxyID, proxyStats.ActiveConns))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_latency_seconds_avg Average latency for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_latency_seconds_avg gauge\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_latency_seconds_avg{proxy_id="%s"} %f\n\n`, proxyID, proxyStats.LatencyAvg.Seconds()))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_latency_seconds_p50 P50 latency for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_latency_seconds_p50 gauge\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_latency_seconds_p50{proxy_id="%s"} %f\n\n`, proxyID, proxyStats.LatencyP50.Seconds()))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_latency_seconds_p95 P95 latency for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_latency_seconds_p95 gauge\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_latency_seconds_p95{proxy_id="%s"} %f\n\n`, proxyID, proxyStats.LatencyP95.Seconds()))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_latency_seconds_p99 P99 latency for proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_latency_seconds_p99 gauge\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_latency_seconds_p99{proxy_id="%s"} %f\n\n`, proxyID, proxyStats.LatencyP99.Seconds()))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_bytes_in_total Total bytes received by proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_bytes_in_total counter\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_bytes_in_total{proxy_id="%s"} %d\n\n`, proxyID, proxyStats.BytesIn))

		output.WriteString(fmt.Sprintf("# HELP modbridge_proxy_bytes_out_total Total bytes sent by proxy\n"))
		output.WriteString(fmt.Sprintf("# TYPE modbridge_proxy_bytes_out_total counter\n"))
		output.WriteString(fmt.Sprintf(`modbridge_proxy_bytes_out_total{proxy_id="%s"} %d\n\n`, proxyID, proxyStats.BytesOut))
	}

	return output.String()
}
