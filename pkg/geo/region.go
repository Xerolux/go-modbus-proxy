package geo

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// RegionManager manages multi-region deployments.
type RegionManager struct {
	mu sync.RWMutex

	localRegion string
	etcdClient  *clientv3.Client

	regions map[string]*Region // region name -> Region
	proxies map[string]string  // proxy ID -> region name
}

// Region represents a geographic region.
type Region struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Location    GeoLocation       `json:"location"`
	Endpoints   []string          `json:"endpoints"` // API endpoints for this region
	Metadata    map[string]string `json:"metadata"`

	// Health and metrics
	IsHealthy      bool      `json:"is_healthy"`
	LastHealthCheck time.Time `json:"last_health_check"`
	AverageLatency  float64   `json:"average_latency_ms"` // Average latency to this region
	ProxyCount     int       `json:"proxy_count"`
}

// GeoLocation represents geographic coordinates.
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
}

// NewRegionManager creates a new region manager.
func NewRegionManager(localRegion string, etcdClient *clientv3.Client) *RegionManager {
	return &RegionManager{
		localRegion: localRegion,
		etcdClient:  etcdClient,
		regions:     make(map[string]*Region),
		proxies:     make(map[string]string),
	}
}

// RegisterRegion registers a new region.
func (rm *RegionManager) RegisterRegion(ctx context.Context, region *Region) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.regions[region.Name]; exists {
		return fmt.Errorf("region already registered: %s", region.Name)
	}

	rm.regions[region.Name] = region

	// Store in etcd for cross-region discovery
	if rm.etcdClient != nil {
		data, err := jsonMarshal(region)
		if err != nil {
			return err
		}

		key := fmt.Sprintf("/modbridge/regions/%s", region.Name)
		_, err = rm.etcdClient.Put(ctx, key, string(data))
		if err != nil {
			return fmt.Errorf("failed to register region in etcd: %w", err)
		}
	}

	return nil
}

// GetRegion retrieves a region by name.
func (rm *RegionManager) GetRegion(name string) (*Region, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	region, ok := rm.regions[name]
	if !ok {
		return nil, fmt.Errorf("region not found: %s", name)
	}

	return region, nil
}

// ListRegions returns all registered regions.
func (rm *RegionManager) ListRegions() []*Region {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	regions := make([]*Region, 0, len(rm.regions))
	for _, region := range rm.regions {
		regions = append(regions, region)
	}

	return regions
}

// GetLocalRegion returns the local region name.
func (rm *RegionManager) GetLocalRegion() string {
	return rm.localRegion
}

// AssignProxyToRegion assigns a proxy to a region.
func (rm *RegionManager) AssignProxyToRegion(proxyID, regionName string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, ok := rm.regions[regionName]; !ok {
		return fmt.Errorf("region not found: %s", regionName)
	}

	rm.proxies[proxyID] = regionName
	rm.regions[regionName].ProxyCount++

	return nil
}

// GetProxyRegion returns the region for a proxy.
func (rm *RegionManager) GetProxyRegion(proxyID string) (string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	region, ok := rm.proxies[proxyID]
	if !ok {
		return "", fmt.Errorf("proxy not assigned to any region: %s", proxyID)
	}

	return region, nil
}

// FindClosestRegion finds the closest region to a given location.
func (rm *RegionManager) FindClosestRegion(location GeoLocation) (*Region, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.regions) == 0 {
		return nil, fmt.Errorf("no regions registered")
	}

	var closest *Region
	minDistance := float64(-1)

	for _, region := range rm.regions {
		if !region.IsHealthy {
			continue // Skip unhealthy regions
		}

		distance := calculateDistance(location, region.Location)
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			closest = region
		}
	}

	if closest == nil {
		return nil, fmt.Errorf("no healthy regions available")
	}

	return closest, nil
}

// UpdateRegionHealth updates the health status of a region.
func (rm *RegionManager) UpdateRegionHealth(regionName string, healthy bool, latency float64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	region, ok := rm.regions[regionName]
	if !ok {
		return fmt.Errorf("region not found: %s", regionName)
	}

	region.IsHealthy = healthy
	region.LastHealthCheck = time.Now()
	region.AverageLatency = latency

	return nil
}

// calculateDistance calculates the Haversine distance between two locations in kilometers.
func calculateDistance(loc1, loc2 GeoLocation) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	lat1 := toRadians(loc1.Latitude)
	lat2 := toRadians(loc2.Latitude)
	deltaLat := toRadians(loc2.Latitude - loc1.Latitude)
	deltaLon := toRadians(loc2.Longitude - loc1.Longitude)

	a := sin(deltaLat/2)*sin(deltaLat/2) +
		cos(lat1)*cos(lat2)*sin(deltaLon/2)*sin(deltaLon/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return earthRadius * c
}

// Helper functions for distance calculation
func toRadians(degrees float64) float64 {
	return degrees * 3.14159265359 / 180.0
}

func sin(x float64) float64 {
	// Simple sin approximation (Taylor series)
	// For production, use math.Sin
	return x - (x*x*x)/6.0 + (x*x*x*x*x)/120.0
}

func cos(x float64) float64 {
	// Simple cos approximation
	// For production, use math.Cos
	return 1.0 - (x*x)/2.0 + (x*x*x*x)/24.0
}

func sqrt(x float64) float64 {
	// Newton's method for square root
	// For production, use math.Sqrt
	if x < 0 {
		return 0
	}
	z := 1.0
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func atan2(y, x float64) float64 {
	// Simplified atan2
	// For production, use math.Atan2
	if x > 0 {
		return atan(y / x)
	}
	return 1.5708 // π/2 approximation
}

func atan(x float64) float64 {
	// Simple atan approximation
	// For production, use math.Atan
	return x / (1.0 + 0.28*x*x)
}

func jsonMarshal(v interface{}) ([]byte, error) {
	// Placeholder - would use encoding/json
	return []byte(fmt.Sprintf("%v", v)), nil
}

// Common region definitions
var (
	RegionUSEast1 = &Region{
		Name:        "us-east-1",
		DisplayName: "US East (Virginia)",
		Location: GeoLocation{
			Latitude:  38.9072,
			Longitude: -77.0369,
			City:      "Ashburn",
			Country:   "US",
		},
	}

	RegionUSWest1 = &Region{
		Name:        "us-west-1",
		DisplayName: "US West (California)",
		Location: GeoLocation{
			Latitude:  37.3541,
			Longitude: -121.9552,
			City:      "San Jose",
			Country:   "US",
		},
	}

	RegionEUWest1 = &Region{
		Name:        "eu-west-1",
		DisplayName: "EU West (Ireland)",
		Location: GeoLocation{
			Latitude:  53.3498,
			Longitude: -6.2603,
			City:      "Dublin",
			Country:   "IE",
		},
	}

	RegionAPSoutheast1 = &Region{
		Name:        "ap-southeast-1",
		DisplayName: "Asia Pacific (Singapore)",
		Location: GeoLocation{
			Latitude:  1.3521,
			Longitude: 103.8198,
			City:      "Singapore",
			Country:   "SG",
		},
	}
)

// LatencyBasedRouter routes requests to the lowest-latency region.
type LatencyBasedRouter struct {
	mu sync.RWMutex

	regionManager *RegionManager
	latencyCache  map[string]time.Duration // client IP -> best region latency
	cacheTTL      time.Duration
}

// NewLatencyBasedRouter creates a new latency-based router.
func NewLatencyBasedRouter(rm *RegionManager) *LatencyBasedRouter {
	return &LatencyBasedRouter{
		regionManager: rm,
		latencyCache:  make(map[string]time.Duration),
		cacheTTL:      5 * time.Minute,
	}
}

// RouteRequest routes a request to the best region based on latency.
func (lr *LatencyBasedRouter) RouteRequest(clientIP string) (*Region, error) {
	regions := lr.regionManager.ListRegions()
	if len(regions) == 0 {
		return nil, fmt.Errorf("no regions available")
	}

	// Find the region with lowest latency
	var bestRegion *Region
	var lowestLatency float64 = -1

	for _, region := range regions {
		if !region.IsHealthy {
			continue
		}

		// Use cached average latency
		if lowestLatency < 0 || region.AverageLatency < lowestLatency {
			lowestLatency = region.AverageLatency
			bestRegion = region
		}
	}

	if bestRegion == nil {
		return nil, fmt.Errorf("no healthy regions available")
	}

	return bestRegion, nil
}

// MeasureLatency measures latency to a region.
func (lr *LatencyBasedRouter) MeasureLatency(regionName string) (time.Duration, error) {
	region, err := lr.regionManager.GetRegion(regionName)
	if err != nil {
		return 0, err
	}

	if len(region.Endpoints) == 0 {
		return 0, fmt.Errorf("no endpoints defined for region %s", regionName)
	}

	// Ping the first endpoint
	endpoint := region.Endpoints[0]

	start := time.Now()
	conn, err := net.DialTimeout("tcp", endpoint, 5*time.Second)
	if err != nil {
		return 0, err
	}
	conn.Close()

	latency := time.Since(start)

	// Update region latency
	lr.regionManager.UpdateRegionHealth(regionName, true, float64(latency.Milliseconds()))

	return latency, nil
}
