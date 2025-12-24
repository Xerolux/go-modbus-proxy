package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Handler provides HTTP handlers for metrics endpoints.
type Handler struct {
	collector *Collector
}

// NewHandler creates a new metrics handler.
func NewHandler(collector *Collector) *Handler {
	return &Handler{
		collector: collector,
	}
}

// HandleMetrics returns a summary of all metrics.
func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := h.collector.Snapshot()

	response := MetricsResponse{
		Uptime:              formatDuration(snapshot.Uptime),
		TotalRequests:       snapshot.TotalRequests,
		TotalErrors:         snapshot.TotalErrors,
		ErrorRate:           snapshot.ErrorRate,
		ActiveConnections:   snapshot.ActiveConnections,
		TotalConnections:    snapshot.TotalConnections,
		RejectedConnections: snapshot.RejectedConnections,
		RequestsPerSec:      snapshot.RequestsPerSec,
		Performance: PerformanceMetrics{
			AvgLatencyMs: float64(snapshot.LatencyStats.Avg) / 1000.0,
			P50LatencyMs: float64(snapshot.LatencyStats.P50) / 1000.0,
			P95LatencyMs: float64(snapshot.LatencyStats.P95) / 1000.0,
			P99LatencyMs: float64(snapshot.LatencyStats.P99) / 1000.0,
			MinLatencyMs: float64(snapshot.LatencyStats.Min) / 1000.0,
			MaxLatencyMs: float64(snapshot.LatencyStats.Max) / 1000.0,
		},
		Resources: ResourceMetrics{
			Goroutines:    snapshot.RuntimeStats.Goroutines,
			HeapAllocMB:   float64(snapshot.RuntimeStats.HeapAlloc) / 1024 / 1024,
			HeapInuseMB:   float64(snapshot.RuntimeStats.HeapInuse) / 1024 / 1024,
			HeapObjects:   snapshot.RuntimeStats.HeapObjects,
			GCRuns:        snapshot.RuntimeStats.GCRuns,
			LastGCPauseMs: float64(snapshot.RuntimeStats.LastGCPause.Microseconds()) / 1000.0,
			MemoryAllocMB: float64(snapshot.RuntimeStats.MemoryAlloc) / 1024 / 1024,
			MemoryTotalMB: float64(snapshot.RuntimeStats.MemoryTotal) / 1024 / 1024,
			MemorySysMB:   float64(snapshot.RuntimeStats.MemorySys) / 1024 / 1024,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleProxyMetrics returns metrics for a specific proxy.
func (h *Handler) HandleProxyMetrics(w http.ResponseWriter, r *http.Request) {
	proxyID := r.URL.Query().Get("id")
	if proxyID == "" {
		http.Error(w, "proxy id required", http.StatusBadRequest)
		return
	}

	snapshot := h.collector.Snapshot()
	proxyMetrics, ok := snapshot.ProxyMetrics[proxyID]
	if !ok {
		http.Error(w, "proxy not found", http.StatusNotFound)
		return
	}

	response := ProxyMetricsResponse{
		ID:                proxyID,
		Name:              proxyMetrics.Name,
		Requests:          proxyMetrics.Requests,
		Errors:            proxyMetrics.Errors,
		BytesReceived:     proxyMetrics.BytesReceived,
		BytesSent:         proxyMetrics.BytesSent,
		ActiveConnections: proxyMetrics.ActiveConnections,
		TotalConnections:  proxyMetrics.TotalConnections,
		AvgLatencyMs:      float64(proxyMetrics.AvgLatencyUs) / 1000.0,
	}

	if proxyMetrics.Requests > 0 {
		response.ErrorRate = float64(proxyMetrics.Errors) / float64(proxyMetrics.Requests)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAllProxyMetrics returns metrics for all proxies.
func (h *Handler) HandleAllProxyMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := h.collector.Snapshot()

	proxies := make([]ProxyMetricsResponse, 0, len(snapshot.ProxyMetrics))
	for id, pm := range snapshot.ProxyMetrics {
		response := ProxyMetricsResponse{
			ID:                id,
			Name:              pm.Name,
			Requests:          pm.Requests,
			Errors:            pm.Errors,
			BytesReceived:     pm.BytesReceived,
			BytesSent:         pm.BytesSent,
			ActiveConnections: pm.ActiveConnections,
			TotalConnections:  pm.TotalConnections,
			AvgLatencyMs:      float64(pm.AvgLatencyUs) / 1000.0,
		}

		if pm.Requests > 0 {
			response.ErrorRate = float64(pm.Errors) / float64(pm.Requests)
		}

		proxies = append(proxies, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proxies)
}

// MetricsResponse is the JSON response for /api/metrics.
type MetricsResponse struct {
	Uptime              string             `json:"uptime"`
	TotalRequests       uint64             `json:"total_requests"`
	TotalErrors         uint64             `json:"total_errors"`
	ErrorRate           float64            `json:"error_rate"`
	ActiveConnections   int64              `json:"active_connections"`
	TotalConnections    uint64             `json:"total_connections"`
	RejectedConnections uint64             `json:"rejected_connections"`
	RequestsPerSec      float64            `json:"requests_per_sec"`
	Performance         PerformanceMetrics `json:"performance"`
	Resources           ResourceMetrics    `json:"resources"`
}

// PerformanceMetrics holds performance-related metrics.
type PerformanceMetrics struct {
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P50LatencyMs float64 `json:"p50_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`
	MinLatencyMs float64 `json:"min_latency_ms"`
	MaxLatencyMs float64 `json:"max_latency_ms"`
}

// ResourceMetrics holds resource usage metrics.
type ResourceMetrics struct {
	Goroutines    int     `json:"goroutines"`
	HeapAllocMB   float64 `json:"heap_alloc_mb"`
	HeapInuseMB   float64 `json:"heap_inuse_mb"`
	HeapObjects   uint64  `json:"heap_objects"`
	GCRuns        uint32  `json:"gc_runs"`
	LastGCPauseMs float64 `json:"last_gc_pause_ms"`
	MemoryAllocMB float64 `json:"memory_alloc_mb"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	MemorySysMB   float64 `json:"memory_sys_mb"`
}

// ProxyMetricsResponse is the JSON response for proxy metrics.
type ProxyMetricsResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Requests          uint64  `json:"requests"`
	Errors            uint64  `json:"errors"`
	ErrorRate         float64 `json:"error_rate"`
	BytesReceived     uint64  `json:"bytes_received"`
	BytesSent         uint64  `json:"bytes_sent"`
	ActiveConnections int64   `json:"active_connections"`
	TotalConnections  uint64  `json:"total_connections"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
}

// formatDuration formats a duration in human-readable form.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
