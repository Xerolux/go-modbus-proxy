package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// ReplicationManager manages data replication across regions.
type ReplicationManager struct {
	mu sync.RWMutex

	etcdClient    *clientv3.Client
	localRegion   string
	regionManager *RegionManager

	// Replication state
	replicas      map[string]*ReplicaInfo // region name -> replica info
	syncInterval  time.Duration
	stopChan      chan struct{}
}

// ReplicaInfo holds information about a replica.
type ReplicaInfo struct {
	Region       string    `json:"region"`
	Endpoint     string    `json:"endpoint"`
	Status       string    `json:"status"` // active, lagging, disconnected
	LastSync     time.Time `json:"last_sync"`
	Lag          time.Duration `json:"lag"`
	BytesSynced  uint64    `json:"bytes_synced"`
	RecordsSynced uint64   `json:"records_synced"`
}

// ReplicationStrategy defines how data is replicated.
type ReplicationStrategy string

const (
	// ReplicationSync means write to all regions synchronously.
	ReplicationSync ReplicationStrategy = "sync"

	// ReplicationAsync means write locally first, replicate asynchronously.
	ReplicationAsync ReplicationStrategy = "async"

	// ReplicationQuorum means write to quorum of regions.
	ReplicationQuorum ReplicationStrategy = "quorum"
)

// ReplicationConfig holds replication configuration.
type ReplicationConfig struct {
	Enabled       bool                `json:"enabled"`
	Strategy      ReplicationStrategy `json:"strategy"`
	SyncInterval  time.Duration       `json:"sync_interval"`
	Regions       []string            `json:"regions"` // Regions to replicate to
	QuorumSize    int                 `json:"quorum_size"` // For quorum strategy
}

// DefaultReplicationConfig returns default replication configuration.
func DefaultReplicationConfig() ReplicationConfig {
	return ReplicationConfig{
		Enabled:      false,
		Strategy:     ReplicationAsync,
		SyncInterval: 10 * time.Second,
		Regions:      []string{},
		QuorumSize:   2,
	}
}

// NewReplicationManager creates a new replication manager.
func NewReplicationManager(etcdClient *clientv3.Client, localRegion string, rm *RegionManager) *ReplicationManager {
	return &ReplicationManager{
		etcdClient:    etcdClient,
		localRegion:   localRegion,
		regionManager: rm,
		replicas:      make(map[string]*ReplicaInfo),
		syncInterval:  10 * time.Second,
		stopChan:      make(chan struct{}),
	}
}

// Start starts the replication manager.
func (rm *ReplicationManager) Start(ctx context.Context) {
	log.Println("[Replication] Starting replication manager")

	ticker := time.NewTicker(rm.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.syncReplicas(ctx)
		}
	}
}

// Stop stops the replication manager.
func (rm *ReplicationManager) Stop() {
	close(rm.stopChan)
}

// AddReplica adds a replica region.
func (rm *ReplicationManager) AddReplica(region, endpoint string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if region == rm.localRegion {
		return fmt.Errorf("cannot add local region as replica")
	}

	rm.replicas[region] = &ReplicaInfo{
		Region:   region,
		Endpoint: endpoint,
		Status:   "active",
		LastSync: time.Now(),
	}

	log.Printf("[Replication] Added replica: %s (%s)", region, endpoint)
	return nil
}

// RemoveReplica removes a replica region.
func (rm *ReplicationManager) RemoveReplica(region string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, ok := rm.replicas[region]; !ok {
		return fmt.Errorf("replica not found: %s", region)
	}

	delete(rm.replicas, region)
	log.Printf("[Replication] Removed replica: %s", region)
	return nil
}

// ReplicateData replicates data to other regions.
func (rm *ReplicationManager) ReplicateData(ctx context.Context, key string, data []byte, strategy ReplicationStrategy) error {
	switch strategy {
	case ReplicationSync:
		return rm.syncReplicate(ctx, key, data)
	case ReplicationAsync:
		go rm.asyncReplicate(ctx, key, data)
		return nil
	case ReplicationQuorum:
		return rm.quorumReplicate(ctx, key, data)
	default:
		return fmt.Errorf("unknown replication strategy: %s", strategy)
	}
}

// syncReplicate replicates data synchronously to all regions.
func (rm *ReplicationManager) syncReplicate(ctx context.Context, key string, data []byte) error {
	rm.mu.RLock()
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	rm.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(replicas))

	for _, replica := range replicas {
		wg.Add(1)
		go func(r *ReplicaInfo) {
			defer wg.Done()

			if err := rm.replicateToRegion(ctx, r.Region, key, data); err != nil {
				errChan <- err
			}
		}(replica)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("replication failed to %d regions: %v", len(errors), errors[0])
	}

	return nil
}

// asyncReplicate replicates data asynchronously.
func (rm *ReplicationManager) asyncReplicate(ctx context.Context, key string, data []byte) {
	rm.mu.RLock()
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	rm.mu.RUnlock()

	for _, replica := range replicas {
		if err := rm.replicateToRegion(ctx, replica.Region, key, data); err != nil {
			log.Printf("[Replication] Failed to replicate to %s: %v", replica.Region, err)
		}
	}
}

// quorumReplicate replicates data to a quorum of regions.
func (rm *ReplicationManager) quorumReplicate(ctx context.Context, key string, data []byte) error {
	rm.mu.RLock()
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	quorumSize := (len(replicas) / 2) + 1 // Majority
	rm.mu.RUnlock()

	if len(replicas) == 0 {
		return nil // No replicas to write to
	}

	successChan := make(chan bool, len(replicas))
	var wg sync.WaitGroup

	for _, replica := range replicas {
		wg.Add(1)
		go func(r *ReplicaInfo) {
			defer wg.Done()

			if err := rm.replicateToRegion(ctx, r.Region, key, data); err != nil {
				log.Printf("[Replication] Failed to replicate to %s: %v", r.Region, err)
				successChan <- false
			} else {
				successChan <- true
			}
		}(replica)
	}

	// Wait for quorum or all to complete
	go func() {
		wg.Wait()
		close(successChan)
	}()

	successCount := 0
	for success := range successChan {
		if success {
			successCount++
		}
		if successCount >= quorumSize {
			return nil // Quorum reached
		}
	}

	if successCount < quorumSize {
		return fmt.Errorf("failed to reach quorum: only %d/%d replicas succeeded", successCount, quorumSize)
	}

	return nil
}

// replicateToRegion replicates data to a specific region.
func (rm *ReplicationManager) replicateToRegion(ctx context.Context, region, key string, data []byte) error {
	// In a real implementation, this would:
	// 1. Connect to the remote region's etcd or API
	// 2. Write the data
	// 3. Verify the write

	// For now, we'll just log it
	log.Printf("[Replication] Replicating %d bytes to region %s (key: %s)", len(data), region, key)

	// Update replica info
	rm.mu.Lock()
	if replica, ok := rm.replicas[region]; ok {
		replica.LastSync = time.Now()
		replica.BytesSynced += uint64(len(data))
		replica.RecordsSynced++
		replica.Status = "active"
	}
	rm.mu.Unlock()

	return nil
}

// syncReplicas periodically syncs with replicas.
func (rm *ReplicationManager) syncReplicas(ctx context.Context) {
	rm.mu.RLock()
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	rm.mu.RUnlock()

	for _, replica := range replicas {
		// Check replica lag
		lag := time.Since(replica.LastSync)

		rm.mu.Lock()
		replica.Lag = lag

		// Update status based on lag
		if lag > 1*time.Minute {
			replica.Status = "lagging"
		}
		if lag > 5*time.Minute {
			replica.Status = "disconnected"
		}
		rm.mu.Unlock()
	}
}

// GetReplicationStatus returns the current replication status.
func (rm *ReplicationManager) GetReplicationStatus() ReplicationStatus {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	replicas := make([]ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, *replica)
	}

	activeCount := 0
	laggingCount := 0
	disconnectedCount := 0

	for _, replica := range replicas {
		switch replica.Status {
		case "active":
			activeCount++
		case "lagging":
			laggingCount++
		case "disconnected":
			disconnectedCount++
		}
	}

	return ReplicationStatus{
		LocalRegion:       rm.localRegion,
		Replicas:          replicas,
		ActiveReplicas:    activeCount,
		LaggingReplicas:   laggingCount,
		DisconnectedReplicas: disconnectedCount,
	}
}

// ReplicationStatus represents the current replication status.
type ReplicationStatus struct {
	LocalRegion          string
	Replicas             []ReplicaInfo
	ActiveReplicas       int
	LaggingReplicas      int
	DisconnectedReplicas int
}

// ConflictResolver resolves data conflicts during replication.
type ConflictResolver struct {
	strategy ConflictResolutionStrategy
}

// ConflictResolutionStrategy defines how conflicts are resolved.
type ConflictResolutionStrategy string

const (
	// LastWriteWins uses timestamp to resolve conflicts.
	LastWriteWins ConflictResolutionStrategy = "last_write_wins"

	// LocalPreference prefers local writes over remote.
	LocalPreference ConflictResolutionStrategy = "local_preference"

	// CustomResolver uses custom resolution logic.
	CustomResolver ConflictResolutionStrategy = "custom"
)

// ResolveConflict resolves a conflict between two versions.
func (cr *ConflictResolver) ResolveConflict(local, remote *VersionedData) *VersionedData {
	switch cr.strategy {
	case LastWriteWins:
		if remote.Timestamp.After(local.Timestamp) {
			return remote
		}
		return local

	case LocalPreference:
		return local

	default:
		// Default to last write wins
		if remote.Timestamp.After(local.Timestamp) {
			return remote
		}
		return local
	}
}

// VersionedData represents data with versioning information.
type VersionedData struct {
	Key       string
	Data      []byte
	Timestamp time.Time
	Region    string
	Version   uint64
}

// MarshalVersionedData marshals versioned data to JSON.
func MarshalVersionedData(vd *VersionedData) ([]byte, error) {
	return json.Marshal(vd)
}

// UnmarshalVersionedData unmarshals versioned data from JSON.
func UnmarshalVersionedData(data []byte) (*VersionedData, error) {
	var vd VersionedData
	if err := json.Unmarshal(data, &vd); err != nil {
		return nil, err
	}
	return &vd, nil
}
