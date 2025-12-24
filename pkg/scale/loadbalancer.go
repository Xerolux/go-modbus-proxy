package scale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// LoadBalancerIntegration provides integration with external load balancers.
// Supports health checks, readiness probes, and graceful shutdown.
type LoadBalancerIntegration struct {
	mu sync.RWMutex

	config      LoadBalancerConfig
	isHealthy   bool
	isReady     bool
	shutdownMode bool

	// Health check state
	lastHealthCheck time.Time
	healthCheckCount uint64
	failedHealthChecks uint64

	// Traffic management
	activeConnections int64
	drainingStartTime time.Time
}

// LoadBalancerConfig holds load balancer configuration.
type LoadBalancerConfig struct {
	HealthCheckEnabled  bool          `json:"health_check_enabled"`
	HealthCheckPath     string        `json:"health_check_path"`      // Default: /health
	ReadinessCheckPath  string        `json:"readiness_check_path"`   // Default: /ready
	DrainTimeout        time.Duration `json:"drain_timeout"`          // Time to wait for connections to drain
	GracefulShutdownEnabled bool      `json:"graceful_shutdown_enabled"`
}

// DefaultLoadBalancerConfig returns default load balancer configuration.
func DefaultLoadBalancerConfig() LoadBalancerConfig {
	return LoadBalancerConfig{
		HealthCheckEnabled:      true,
		HealthCheckPath:         "/health",
		ReadinessCheckPath:      "/ready",
		DrainTimeout:            30 * time.Second,
		GracefulShutdownEnabled: true,
	}
}

// NewLoadBalancerIntegration creates a new load balancer integration.
func NewLoadBalancerIntegration(config LoadBalancerConfig) *LoadBalancerIntegration {
	return &LoadBalancerIntegration{
		config:    config,
		isHealthy: true,
		isReady:   true,
	}
}

// RegisterHandlers registers HTTP handlers for health checks.
func (lb *LoadBalancerIntegration) RegisterHandlers(mux *http.ServeMux) {
	if lb.config.HealthCheckEnabled {
		mux.HandleFunc(lb.config.HealthCheckPath, lb.HealthCheckHandler)
		mux.HandleFunc(lb.config.ReadinessCheckPath, lb.ReadinessCheckHandler)
	}

	log.Printf("[LoadBalancer] Registered health check endpoints: %s, %s",
		lb.config.HealthCheckPath, lb.config.ReadinessCheckPath)
}

// HealthCheckHandler handles health check requests.
// Returns 200 if healthy, 503 if unhealthy.
// Health checks verify that the service is running and can handle traffic.
func (lb *LoadBalancerIntegration) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	lb.mu.Lock()
	lb.lastHealthCheck = time.Now()
	lb.healthCheckCount++
	lb.mu.Unlock()

	lb.mu.RLock()
	healthy := lb.isHealthy
	lb.mu.RUnlock()

	if healthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	} else {
		lb.mu.Lock()
		lb.failedHealthChecks++
		lb.mu.Unlock()

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy"}`))
	}
}

// ReadinessCheckHandler handles readiness check requests.
// Returns 200 if ready to accept traffic, 503 if not ready.
// Readiness checks verify that the service is ready to handle requests.
func (lb *LoadBalancerIntegration) ReadinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	lb.mu.RLock()
	ready := lb.isReady && !lb.shutdownMode
	activeConns := lb.activeConnections
	lb.mu.RUnlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","active_connections":%d}`, activeConns)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","active_connections":%d}`, activeConns)
	}
}

// SetHealthy sets the health status.
func (lb *LoadBalancerIntegration) SetHealthy(healthy bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.isHealthy != healthy {
		log.Printf("[LoadBalancer] Health status changed: %v -> %v", lb.isHealthy, healthy)
		lb.isHealthy = healthy
	}
}

// SetReady sets the readiness status.
func (lb *LoadBalancerIntegration) SetReady(ready bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.isReady != ready {
		log.Printf("[LoadBalancer] Readiness status changed: %v -> %v", lb.isReady, ready)
		lb.isReady = ready
	}
}

// RecordConnection records a new active connection.
func (lb *LoadBalancerIntegration) RecordConnection() {
	lb.mu.Lock()
	lb.activeConnections++
	lb.mu.Unlock()
}

// RecordDisconnection records a connection being closed.
func (lb *LoadBalancerIntegration) RecordDisconnection() {
	lb.mu.Lock()
	lb.activeConnections--
	lb.mu.Unlock()
}

// GetActiveConnections returns the number of active connections.
func (lb *LoadBalancerIntegration) GetActiveConnections() int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.activeConnections
}

// PrepareShutdown prepares for graceful shutdown.
// Sets readiness to false and waits for connections to drain.
func (lb *LoadBalancerIntegration) PrepareShutdown(ctx context.Context) error {
	if !lb.config.GracefulShutdownEnabled {
		log.Println("[LoadBalancer] Graceful shutdown disabled, shutting down immediately")
		return nil
	}

	log.Println("[LoadBalancer] Preparing for graceful shutdown...")

	// Mark as not ready (stops receiving new traffic from LB)
	lb.mu.Lock()
	lb.shutdownMode = true
	lb.isReady = false
	lb.drainingStartTime = time.Now()
	lb.mu.Unlock()

	log.Println("[LoadBalancer] Marked as not ready, waiting for connections to drain...")

	// Wait for connections to drain or timeout
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.NewTimer(lb.config.DrainTimeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timeout.C:
			lb.mu.RLock()
			activeConns := lb.activeConnections
			lb.mu.RUnlock()

			log.Printf("[LoadBalancer] Drain timeout reached, %d connections still active", activeConns)
			return fmt.Errorf("drain timeout: %d connections still active", activeConns)

		case <-ticker.C:
			lb.mu.RLock()
			activeConns := lb.activeConnections
			lb.mu.RUnlock()

			if activeConns == 0 {
				drainDuration := time.Since(lb.drainingStartTime)
				log.Printf("[LoadBalancer] All connections drained in %s", drainDuration)
				return nil
			}

			log.Printf("[LoadBalancer] Waiting for %d connections to drain...", activeConns)
		}
	}
}

// GetStats returns load balancer statistics.
func (lb *LoadBalancerIntegration) GetStats() LoadBalancerStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return LoadBalancerStats{
		Healthy:             lb.isHealthy,
		Ready:               lb.isReady,
		ActiveConnections:   lb.activeConnections,
		HealthCheckCount:    lb.healthCheckCount,
		FailedHealthChecks:  lb.failedHealthChecks,
		LastHealthCheck:     lb.lastHealthCheck,
		ShutdownMode:        lb.shutdownMode,
		DrainStartTime:      lb.drainingStartTime,
	}
}

// LoadBalancerStats holds load balancer statistics.
type LoadBalancerStats struct {
	Healthy            bool
	Ready              bool
	ActiveConnections  int64
	HealthCheckCount   uint64
	FailedHealthChecks uint64
	LastHealthCheck    time.Time
	ShutdownMode       bool
	DrainStartTime     time.Time
}

// HAProxyHealthCheck provides HAProxy-specific health check format.
func (lb *LoadBalancerIntegration) HAProxyHealthCheck(w http.ResponseWriter, r *http.Request) {
	lb.mu.RLock()
	healthy := lb.isHealthy && lb.isReady && !lb.shutdownMode
	lb.mu.RUnlock()

	if healthy {
		// HAProxy expects simple text response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("UNAVAILABLE\n"))
	}
}

// NginxHealthCheck provides NGINX-specific health check format.
func (lb *LoadBalancerIntegration) NginxHealthCheck(w http.ResponseWriter, r *http.Request) {
	// NGINX uses standard HTTP status codes
	lb.HealthCheckHandler(w, r)
}

// KubernetesLivenessProbe provides Kubernetes liveness probe.
// Liveness probes check if the container is alive.
func (lb *LoadBalancerIntegration) KubernetesLivenessProbe(w http.ResponseWriter, r *http.Request) {
	// Liveness should only fail if the process is truly broken
	// Don't fail on temporary issues or during shutdown
	lb.mu.RLock()
	healthy := lb.isHealthy
	lb.mu.RUnlock()

	if healthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy"}`))
	}
}

// KubernetesReadinessProbe provides Kubernetes readiness probe.
// Readiness probes check if the container can accept traffic.
func (lb *LoadBalancerIntegration) KubernetesReadinessProbe(w http.ResponseWriter, r *http.Request) {
	lb.ReadinessCheckHandler(w, r)
}

// AWSTargetGroupHealthCheck provides AWS ALB/NLB health check format.
func (lb *LoadBalancerIntegration) AWSTargetGroupHealthCheck(w http.ResponseWriter, r *http.Request) {
	lb.mu.RLock()
	healthy := lb.isHealthy && lb.isReady && !lb.shutdownMode
	lb.mu.RUnlock()

	if healthy {
		w.WriteHeader(http.StatusOK)
		// AWS expects 200 OK with optional body
		w.Write([]byte(`{"status":"healthy"}`))
	} else {
		// Return 503 to mark target as unhealthy
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy"}`))
	}
}
