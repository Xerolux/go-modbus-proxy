package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// StatelessProxyManager manages proxies in a stateless, horizontally scalable way.
// All state is stored in etcd for multi-instance deployments.
type StatelessProxyManager struct {
	etcdClient  *clientv3.Client
	instanceID  string
	region      string
	mu          sync.RWMutex
	localProxies map[string]*ProxyState // Cached local state
}

// ProxyState represents the state of a proxy instance.
type ProxyState struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	InstanceID    string    `json:"instance_id"` // Which server instance owns this
	Region        string    `json:"region"`
	ListenAddr    string    `json:"listen_addr"`
	TargetAddr    string    `json:"target_addr"`
	Status        string    `json:"status"` // running, stopped, error
	StartedAt     time.Time `json:"started_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Connections   int64     `json:"connections"`
	Requests      uint64    `json:"requests"`
	Errors        uint64    `json:"errors"`
}

// NewStatelessProxyManager creates a new stateless proxy manager.
func NewStatelessProxyManager(etcdClient *clientv3.Client, instanceID, region string) *StatelessProxyManager {
	return &StatelessProxyManager{
		etcdClient:   etcdClient,
		instanceID:   instanceID,
		region:       region,
		localProxies: make(map[string]*ProxyState),
	}
}

// RegisterProxy registers a proxy in the distributed state.
func (m *StatelessProxyManager) RegisterProxy(ctx context.Context, state *ProxyState) error {
	state.InstanceID = m.instanceID
	state.Region = m.region
	state.LastHeartbeat = time.Now()

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal proxy state: %w", err)
	}

	key := fmt.Sprintf("/modbridge/proxies/%s", state.ID)

	// Store in etcd with TTL (auto-cleanup on instance crash)
	lease, err := m.etcdClient.Grant(ctx, 30) // 30-second TTL
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	_, err = m.etcdClient.Put(ctx, key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register proxy: %w", err)
	}

	// Cache locally
	m.mu.Lock()
	m.localProxies[state.ID] = state
	m.mu.Unlock()

	// Start heartbeat goroutine
	go m.heartbeat(ctx, state.ID, lease.ID)

	return nil
}

// UnregisterProxy removes a proxy from the distributed state.
func (m *StatelessProxyManager) UnregisterProxy(ctx context.Context, proxyID string) error {
	key := fmt.Sprintf("/modbridge/proxies/%s", proxyID)

	_, err := m.etcdClient.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister proxy: %w", err)
	}

	m.mu.Lock()
	delete(m.localProxies, proxyID)
	m.mu.Unlock()

	return nil
}

// GetProxyState retrieves the current state of a proxy.
func (m *StatelessProxyManager) GetProxyState(ctx context.Context, proxyID string) (*ProxyState, error) {
	key := fmt.Sprintf("/modbridge/proxies/%s", proxyID)

	resp, err := m.etcdClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy state: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("proxy not found")
	}

	var state ProxyState
	if err := json.Unmarshal(resp.Kvs[0].Value, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proxy state: %w", err)
	}

	return &state, nil
}

// ListProxies returns all proxies across all instances.
func (m *StatelessProxyManager) ListProxies(ctx context.Context) ([]ProxyState, error) {
	resp, err := m.etcdClient.Get(ctx, "/modbridge/proxies/", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list proxies: %w", err)
	}

	proxies := make([]ProxyState, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var state ProxyState
		if err := json.Unmarshal(kv.Value, &state); err != nil {
			continue // Skip invalid entries
		}
		proxies = append(proxies, state)
	}

	return proxies, nil
}

// ListLocalProxies returns proxies running on this instance.
func (m *StatelessProxyManager) ListLocalProxies(ctx context.Context) []ProxyState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proxies := make([]ProxyState, 0, len(m.localProxies))
	for _, state := range m.localProxies {
		proxies = append(proxies, *state)
	}

	return proxies
}

// UpdateProxyMetrics updates proxy metrics in distributed state.
func (m *StatelessProxyManager) UpdateProxyMetrics(ctx context.Context, proxyID string, connections int64, requests, errors uint64) error {
	m.mu.Lock()
	state, ok := m.localProxies[proxyID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("proxy not found locally")
	}

	state.Connections = connections
	state.Requests = requests
	state.Errors = errors
	state.LastHeartbeat = time.Now()
	m.mu.Unlock()

	// Update in etcd
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal proxy state: %w", err)
	}

	key := fmt.Sprintf("/modbridge/proxies/%s", proxyID)
	_, err = m.etcdClient.Put(ctx, key, string(data))
	return err
}

// heartbeat maintains the proxy's presence in etcd.
func (m *StatelessProxyManager) heartbeat(ctx context.Context, proxyID string, leaseID clientv3.LeaseID) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Keep the lease alive
			_, err := m.etcdClient.KeepAliveOnce(ctx, leaseID)
			if err != nil {
				// Lease expired or error, proxy will be auto-removed
				return
			}

			// Update metrics
			m.mu.RLock()
			state, ok := m.localProxies[proxyID]
			m.mu.RUnlock()

			if !ok {
				return // Proxy was removed
			}

			// Update heartbeat timestamp
			state.LastHeartbeat = time.Now()
			data, _ := json.Marshal(state)
			key := fmt.Sprintf("/modbridge/proxies/%s", proxyID)
			m.etcdClient.Put(ctx, key, string(data), clientv3.WithLease(leaseID))
		}
	}
}

// WatchProxies watches for proxy changes across the cluster.
func (m *StatelessProxyManager) WatchProxies(ctx context.Context) <-chan ProxyEvent {
	events := make(chan ProxyEvent, 100)

	go func() {
		defer close(events)

		watchChan := m.etcdClient.Watch(ctx, "/modbridge/proxies/", clientv3.WithPrefix())

		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				var state ProxyState
				if err := json.Unmarshal(event.Kv.Value, &state); err != nil {
					continue
				}

				var eventType ProxyEventType
				switch event.Type {
				case clientv3.EventTypePut:
					eventType = ProxyAdded
				case clientv3.EventTypeDelete:
					eventType = ProxyRemoved
				}

				events <- ProxyEvent{
					Type:  eventType,
					Proxy: state,
				}
			}
		}
	}()

	return events
}

// ProxyEvent represents a proxy state change event.
type ProxyEvent struct {
	Type  ProxyEventType
	Proxy ProxyState
}

// ProxyEventType represents the type of proxy event.
type ProxyEventType int

const (
	// ProxyAdded indicates a new proxy was added or updated.
	ProxyAdded ProxyEventType = iota

	// ProxyRemoved indicates a proxy was removed.
	ProxyRemoved
)

// GetClusterHealth returns health information for all instances.
func (m *StatelessProxyManager) GetClusterHealth(ctx context.Context) (ClusterHealth, error) {
	proxies, err := m.ListProxies(ctx)
	if err != nil {
		return ClusterHealth{}, err
	}

	health := ClusterHealth{
		Instances:     make(map[string]InstanceHealth),
		TotalProxies:  len(proxies),
		HealthyProxies: 0,
	}

	now := time.Now()
	for _, proxy := range proxies {
		// Check if proxy is healthy (heartbeat within last 60 seconds)
		isHealthy := now.Sub(proxy.LastHeartbeat) < 60*time.Second

		if isHealthy && proxy.Status == "running" {
			health.HealthyProxies++
		}

		// Aggregate by instance
		instance, ok := health.Instances[proxy.InstanceID]
		if !ok {
			instance = InstanceHealth{
				InstanceID: proxy.InstanceID,
				Region:     proxy.Region,
			}
		}

		instance.ProxyCount++
		if isHealthy {
			instance.HealthyProxies++
		}
		instance.TotalConnections += proxy.Connections
		instance.TotalRequests += proxy.Requests
		instance.TotalErrors += proxy.Errors

		if proxy.LastHeartbeat.After(instance.LastSeen) {
			instance.LastSeen = proxy.LastHeartbeat
		}

		health.Instances[proxy.InstanceID] = instance
	}

	return health, nil
}

// ClusterHealth represents the health of the entire cluster.
type ClusterHealth struct {
	Instances      map[string]InstanceHealth
	TotalProxies   int
	HealthyProxies int
}

// InstanceHealth represents the health of a single instance.
type InstanceHealth struct {
	InstanceID       string
	Region           string
	ProxyCount       int
	HealthyProxies   int
	TotalConnections int64
	TotalRequests    uint64
	TotalErrors      uint64
	LastSeen         time.Time
}
