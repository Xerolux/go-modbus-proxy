package scale

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// AutoScaler automatically scales proxy instances based on metrics.
type AutoScaler struct {
	mu sync.RWMutex

	config   AutoScaleConfig
	manager  *StatelessProxyManager
	enabled  bool
	stopChan chan struct{}

	// Current state
	desiredReplicas int
	actualReplicas  int

	// Metrics history for trend analysis
	metricsHistory []ScaleMetrics
	historySize    int
}

// AutoScaleConfig holds auto-scaling configuration.
type AutoScaleConfig struct {
	Enabled              bool          `json:"enabled"`
	MinReplicas          int           `json:"min_replicas"`
	MaxReplicas          int           `json:"max_replicas"`
	CheckInterval        time.Duration `json:"check_interval"`
	ScaleUpThreshold     float64       `json:"scale_up_threshold"`      // CPU/memory threshold to scale up (0.0-1.0)
	ScaleDownThreshold   float64       `json:"scale_down_threshold"`    // CPU/memory threshold to scale down
	ConnectionsPerProxy  int64         `json:"connections_per_proxy"`   // Target connections per proxy
	RequestsPerSecond    float64       `json:"requests_per_second"`     // Target RPS per proxy
	CooldownPeriod       time.Duration `json:"cooldown_period"`         // Wait time between scaling actions
	StabilizationWindow  time.Duration `json:"stabilization_window"`    // Window for calculating average metrics
}

// DefaultAutoScaleConfig returns default auto-scaling configuration.
func DefaultAutoScaleConfig() AutoScaleConfig {
	return AutoScaleConfig{
		Enabled:              false,
		MinReplicas:          1,
		MaxReplicas:          10,
		CheckInterval:        30 * time.Second,
		ScaleUpThreshold:     0.75,  // Scale up at 75% utilization
		ScaleDownThreshold:   0.25,  // Scale down at 25% utilization
		ConnectionsPerProxy:  1000,  // 1000 connections per proxy
		RequestsPerSecond:    1000.0, // 1000 RPS per proxy
		CooldownPeriod:       5 * time.Minute,
		StabilizationWindow:  5 * time.Minute,
	}
}

// ScaleMetrics holds metrics used for scaling decisions.
type ScaleMetrics struct {
	Timestamp        time.Time
	TotalConnections int64
	TotalRequests    uint64
	TotalProxies     int
	AvgCPU           float64
	AvgMemory        float64
	RequestsPerSec   float64
}

// NewAutoScaler creates a new auto-scaler.
func NewAutoScaler(config AutoScaleConfig, manager *StatelessProxyManager) *AutoScaler {
	return &AutoScaler{
		config:         config,
		manager:        manager,
		enabled:        config.Enabled,
		stopChan:       make(chan struct{}),
		historySize:    10, // Keep last 10 data points
		metricsHistory: make([]ScaleMetrics, 0, 10),
	}
}

// Start starts the auto-scaler.
func (as *AutoScaler) Start(ctx context.Context) {
	if !as.enabled {
		log.Println("[AutoScaler] Disabled, not starting")
		return
	}

	log.Println("[AutoScaler] Starting auto-scaler")
	ticker := time.NewTicker(as.config.CheckInterval)
	defer ticker.Stop()

	lastScaleAction := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-as.stopChan:
			return
		case <-ticker.C:
			// Check if cooldown period has passed
			if time.Since(lastScaleAction) < as.config.CooldownPeriod {
				continue
			}

			// Collect metrics
			metrics, err := as.collectMetrics(ctx)
			if err != nil {
				log.Printf("[AutoScaler] Failed to collect metrics: %v", err)
				continue
			}

			// Add to history
			as.addMetrics(metrics)

			// Make scaling decision
			decision := as.makeScalingDecision(metrics)

			if decision != ScaleNoOp {
				log.Printf("[AutoScaler] Scaling decision: %v (current: %d, desired: %d)",
					decision, as.actualReplicas, as.desiredReplicas)

				// Execute scaling action
				if err := as.executeScaling(ctx, decision); err != nil {
					log.Printf("[AutoScaler] Failed to execute scaling: %v", err)
				} else {
					lastScaleAction = time.Now()
				}
			}
		}
	}
}

// Stop stops the auto-scaler.
func (as *AutoScaler) Stop() {
	close(as.stopChan)
}

// collectMetrics collects current metrics for scaling decisions.
func (as *AutoScaler) collectMetrics(ctx context.Context) (ScaleMetrics, error) {
	health, err := as.manager.GetClusterHealth(ctx)
	if err != nil {
		return ScaleMetrics{}, err
	}

	var totalConnections int64
	var totalRequests, totalErrors uint64

	for _, instance := range health.Instances {
		totalConnections += instance.TotalConnections
		totalRequests += instance.TotalRequests
		totalErrors += instance.TotalErrors
	}

	// Calculate requests per second (simplified - use last interval)
	var requestsPerSec float64
	if len(as.metricsHistory) > 0 {
		lastMetrics := as.metricsHistory[len(as.metricsHistory)-1]
		timeDiff := time.Since(lastMetrics.Timestamp).Seconds()
		if timeDiff > 0 {
			requestsDiff := float64(totalRequests - lastMetrics.TotalRequests)
			requestsPerSec = requestsDiff / timeDiff
		}
	}

	return ScaleMetrics{
		Timestamp:        time.Now(),
		TotalConnections: totalConnections,
		TotalRequests:    totalRequests,
		TotalProxies:     health.TotalProxies,
		RequestsPerSec:   requestsPerSec,
		// TODO: Collect actual CPU/memory metrics
		AvgCPU:    0.0,
		AvgMemory: 0.0,
	}, nil
}

// addMetrics adds metrics to history.
func (as *AutoScaler) addMetrics(metrics ScaleMetrics) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.metricsHistory = append(as.metricsHistory, metrics)

	// Keep only last N metrics
	if len(as.metricsHistory) > as.historySize {
		as.metricsHistory = as.metricsHistory[1:]
	}

	as.actualReplicas = metrics.TotalProxies
}

// makeScalingDecision determines if scaling is needed.
func (as *AutoScaler) makeScalingDecision(metrics ScaleMetrics) ScaleAction {
	as.mu.RLock()
	defer as.mu.RUnlock()

	currentReplicas := metrics.TotalProxies
	if currentReplicas == 0 {
		currentReplicas = 1
	}

	// Calculate average connections per proxy
	avgConnectionsPerProxy := float64(metrics.TotalConnections) / float64(currentReplicas)
	avgRPSPerProxy := metrics.RequestsPerSec / float64(currentReplicas)

	// Determine desired replicas based on connections
	desiredByConnections := int(float64(metrics.TotalConnections) / float64(as.config.ConnectionsPerProxy))
	if desiredByConnections < 1 {
		desiredByConnections = 1
	}

	// Determine desired replicas based on RPS
	desiredByRPS := int(metrics.RequestsPerSec / as.config.RequestsPerSecond)
	if desiredByRPS < 1 {
		desiredByRPS = 1
	}

	// Take the maximum of the two
	desiredReplicas := desiredByConnections
	if desiredByRPS > desiredReplicas {
		desiredReplicas = desiredByRPS
	}

	// Apply min/max bounds
	if desiredReplicas < as.config.MinReplicas {
		desiredReplicas = as.config.MinReplicas
	}
	if desiredReplicas > as.config.MaxReplicas {
		desiredReplicas = as.config.MaxReplicas
	}

	as.desiredReplicas = desiredReplicas

	log.Printf("[AutoScaler] Metrics: proxies=%d, conn=%d (%.1f/proxy), rps=%.1f (%.1f/proxy)",
		currentReplicas, metrics.TotalConnections, avgConnectionsPerProxy,
		metrics.RequestsPerSec, avgRPSPerProxy)

	log.Printf("[AutoScaler] Desired replicas: %d (by conn=%d, by rps=%d)",
		desiredReplicas, desiredByConnections, desiredByRPS)

	// Determine action
	if desiredReplicas > currentReplicas {
		return ScaleUp
	} else if desiredReplicas < currentReplicas {
		// Only scale down if consistently under-utilized
		if as.isConsistentlyUnderUtilized() {
			return ScaleDown
		}
	}

	return ScaleNoOp
}

// isConsistentlyUnderUtilized checks if the system has been under-utilized for a while.
func (as *AutoScaler) isConsistentlyUnderUtilized() bool {
	// Need at least 3 data points
	if len(as.metricsHistory) < 3 {
		return false
	}

	// Check last 3 metrics
	for i := len(as.metricsHistory) - 3; i < len(as.metricsHistory); i++ {
		metrics := as.metricsHistory[i]
		utilization := float64(metrics.TotalConnections) / (float64(metrics.TotalProxies) * float64(as.config.ConnectionsPerProxy))

		if utilization > as.config.ScaleDownThreshold {
			return false
		}
	}

	return true
}

// executeScaling executes the scaling action.
func (as *AutoScaler) executeScaling(ctx context.Context, action ScaleAction) error {
	switch action {
	case ScaleUp:
		return as.scaleUp(ctx)
	case ScaleDown:
		return as.scaleDown(ctx)
	default:
		return nil
	}
}

// scaleUp scales up by adding more proxy instances.
func (as *AutoScaler) scaleUp(ctx context.Context) error {
	// In a real implementation, this would:
	// 1. Request new instances from orchestrator (Kubernetes, Docker Swarm, etc.)
	// 2. Wait for instances to become ready
	// 3. Register new proxies

	log.Printf("[AutoScaler] Scaling UP from %d to %d replicas", as.actualReplicas, as.desiredReplicas)

	// Placeholder: Would integrate with Kubernetes HPA, AWS Auto Scaling, etc.
	// For now, just log the action

	return fmt.Errorf("scale-up not implemented (would scale from %d to %d)",
		as.actualReplicas, as.desiredReplicas)
}

// scaleDown scales down by removing proxy instances.
func (as *AutoScaler) scaleDown(ctx context.Context) error {
	// In a real implementation, this would:
	// 1. Select instance to remove (least loaded)
	// 2. Drain connections gracefully
	// 3. Terminate instance
	// 4. Update load balancer

	log.Printf("[AutoScaler] Scaling DOWN from %d to %d replicas", as.actualReplicas, as.desiredReplicas)

	// Placeholder: Would integrate with orchestrator
	// For now, just log the action

	return fmt.Errorf("scale-down not implemented (would scale from %d to %d)",
		as.actualReplicas, as.desiredReplicas)
}

// ScaleAction represents a scaling action.
type ScaleAction int

const (
	// ScaleNoOp means no scaling is needed.
	ScaleNoOp ScaleAction = iota

	// ScaleUp means scale up is needed.
	ScaleUp

	// ScaleDown means scale down is needed.
	ScaleDown
)

func (a ScaleAction) String() string {
	switch a {
	case ScaleNoOp:
		return "no-op"
	case ScaleUp:
		return "scale-up"
	case ScaleDown:
		return "scale-down"
	default:
		return "unknown"
	}
}

// GetStatus returns the current auto-scaler status.
func (as *AutoScaler) GetStatus() AutoScalerStatus {
	as.mu.RLock()
	defer as.mu.RUnlock()

	var recentMetrics []ScaleMetrics
	if len(as.metricsHistory) > 0 {
		recentMetrics = make([]ScaleMetrics, len(as.metricsHistory))
		copy(recentMetrics, as.metricsHistory)
	}

	return AutoScalerStatus{
		Enabled:         as.enabled,
		DesiredReplicas: as.desiredReplicas,
		ActualReplicas:  as.actualReplicas,
		RecentMetrics:   recentMetrics,
	}
}

// AutoScalerStatus represents the current status of the auto-scaler.
type AutoScalerStatus struct {
	Enabled         bool
	DesiredReplicas int
	ActualReplicas  int
	RecentMetrics   []ScaleMetrics
}
