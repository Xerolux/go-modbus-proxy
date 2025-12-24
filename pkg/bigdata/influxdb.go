package bigdata

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// InfluxDBExporter exports Modbus metrics to InfluxDB for time-series analysis.
type InfluxDBExporter struct {
	mu sync.RWMutex

	config    InfluxDBConfig
	enabled   bool
	batchSize int
	buffer    []DataPoint
	stopChan  chan struct{}

	// Stats
	exported uint64
	errors   uint64
}

// InfluxDBConfig holds InfluxDB configuration.
type InfluxDBConfig struct {
	Enabled      bool          `json:"enabled"`
	URL          string        `json:"url"`           // e.g., http://localhost:8086
	Token        string        `json:"token"`         // API token
	Organization string        `json:"organization"`  // Organization name
	Bucket       string        `json:"bucket"`        // Bucket name
	BatchSize    int           `json:"batch_size"`    // Batch size for writes
	FlushInterval time.Duration `json:"flush_interval"` // Interval to flush buffer
	Timeout      time.Duration `json:"timeout"`       // Write timeout
}

// DefaultInfluxDBConfig returns default InfluxDB configuration.
func DefaultInfluxDBConfig() InfluxDBConfig {
	return InfluxDBConfig{
		Enabled:       false,
		URL:           "http://localhost:8086",
		Bucket:        "modbridge",
		Organization:  "default",
		BatchSize:     100,
		FlushInterval: 10 * time.Second,
		Timeout:       5 * time.Second,
	}
}

// DataPoint represents a single data point for InfluxDB.
type DataPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Timestamp   time.Time
}

// NewInfluxDBExporter creates a new InfluxDB exporter.
func NewInfluxDBExporter(config InfluxDBConfig) *InfluxDBExporter {
	return &InfluxDBExporter{
		config:    config,
		enabled:   config.Enabled,
		batchSize: config.BatchSize,
		buffer:    make([]DataPoint, 0, config.BatchSize),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the InfluxDB exporter.
func (ie *InfluxDBExporter) Start(ctx context.Context) {
	if !ie.enabled {
		log.Println("[InfluxDB] Disabled, not starting")
		return
	}

	log.Printf("[InfluxDB] Starting exporter (URL: %s, Bucket: %s)", ie.config.URL, ie.config.Bucket)

	ticker := time.NewTicker(ie.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ie.flush() // Final flush
			return
		case <-ie.stopChan:
			ie.flush()
			return
		case <-ticker.C:
			ie.flush()
		}
	}
}

// Stop stops the InfluxDB exporter.
func (ie *InfluxDBExporter) Stop() {
	close(ie.stopChan)
}

// ExportModbusRequest exports a Modbus request to InfluxDB.
func (ie *InfluxDBExporter) ExportModbusRequest(proxyID, proxyName string, functionCode uint8, address, count uint16, latencyMs float64, success bool) {
	if !ie.enabled {
		return
	}

	point := DataPoint{
		Measurement: "modbus_requests",
		Tags: map[string]string{
			"proxy_id":      proxyID,
			"proxy_name":    proxyName,
			"function_code": fmt.Sprintf("0x%02X", functionCode),
			"success":       fmt.Sprintf("%t", success),
		},
		Fields: map[string]interface{}{
			"address":    address,
			"count":      count,
			"latency_ms": latencyMs,
			"value":      1, // Count
		},
		Timestamp: time.Now(),
	}

	ie.addPoint(point)
}

// ExportConnectionMetrics exports connection metrics.
func (ie *InfluxDBExporter) ExportConnectionMetrics(proxyID, proxyName string, activeConnections int64, totalConnections uint64) {
	if !ie.enabled {
		return
	}

	point := DataPoint{
		Measurement: "modbus_connections",
		Tags: map[string]string{
			"proxy_id":   proxyID,
			"proxy_name": proxyName,
		},
		Fields: map[string]interface{}{
			"active_connections": activeConnections,
			"total_connections":  totalConnections,
		},
		Timestamp: time.Now(),
	}

	ie.addPoint(point)
}

// ExportSystemMetrics exports system-level metrics.
func (ie *InfluxDBExporter) ExportSystemMetrics(goroutines int, heapMB, cpuPercent float64) {
	if !ie.enabled {
		return
	}

	point := DataPoint{
		Measurement: "modbus_system",
		Tags: map[string]string{
			"host": "localhost", // TODO: Get actual hostname
		},
		Fields: map[string]interface{}{
			"goroutines":  goroutines,
			"heap_mb":     heapMB,
			"cpu_percent": cpuPercent,
		},
		Timestamp: time.Now(),
	}

	ie.addPoint(point)
}

// ExportDeviceMetrics exports device-specific metrics.
func (ie *InfluxDBExporter) ExportDeviceMetrics(deviceID, proxyID string, registerAddress uint16, value interface{}) {
	if !ie.enabled {
		return
	}

	point := DataPoint{
		Measurement: "modbus_device_data",
		Tags: map[string]string{
			"device_id": deviceID,
			"proxy_id":  proxyID,
			"register":  fmt.Sprintf("%d", registerAddress),
		},
		Fields: map[string]interface{}{
			"value": value,
		},
		Timestamp: time.Now(),
	}

	ie.addPoint(point)
}

// addPoint adds a data point to the buffer.
func (ie *InfluxDBExporter) addPoint(point DataPoint) {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	ie.buffer = append(ie.buffer, point)

	// Flush if buffer is full
	if len(ie.buffer) >= ie.batchSize {
		go ie.flush()
	}
}

// flush writes buffered data points to InfluxDB.
func (ie *InfluxDBExporter) flush() {
	ie.mu.Lock()
	if len(ie.buffer) == 0 {
		ie.mu.Unlock()
		return
	}

	// Swap buffer
	points := ie.buffer
	ie.buffer = make([]DataPoint, 0, ie.batchSize)
	ie.mu.Unlock()

	// Write points to InfluxDB
	if err := ie.writePoints(points); err != nil {
		log.Printf("[InfluxDB] Failed to write %d points: %v", len(points), err)
		ie.mu.Lock()
		ie.errors++
		ie.mu.Unlock()
	} else {
		ie.mu.Lock()
		ie.exported += uint64(len(points))
		ie.mu.Unlock()
	}
}

// writePoints writes data points to InfluxDB.
func (ie *InfluxDBExporter) writePoints(points []DataPoint) error {
	// In a real implementation, this would:
	// 1. Use influxdb-client-go library
	// 2. Create write API
	// 3. Write points in batch
	// 4. Handle retries

	// For now, just log
	log.Printf("[InfluxDB] Would write %d points to bucket '%s'", len(points), ie.config.Bucket)

	// Example of what the real implementation would look like:
	/*
		client := influxdb2.NewClient(ie.config.URL, ie.config.Token)
		writeAPI := client.WriteAPIBlocking(ie.config.Organization, ie.config.Bucket)

		for _, point := range points {
			p := influxdb2.NewPoint(
				point.Measurement,
				point.Tags,
				point.Fields,
				point.Timestamp,
			)
			if err := writeAPI.WritePoint(context.Background(), p); err != nil {
				return err
			}
		}
	*/

	return nil
}

// GetStats returns exporter statistics.
func (ie *InfluxDBExporter) GetStats() (exported, errors uint64) {
	ie.mu.RLock()
	defer ie.mu.RUnlock()
	return ie.exported, ie.errors
}

// QueryBuilder helps build InfluxDB queries.
type QueryBuilder struct {
	bucket      string
	measurement string
	start       time.Time
	stop        time.Time
	filters     map[string]string
	aggregation string
	window      string
}

// NewQueryBuilder creates a new query builder.
func NewQueryBuilder(bucket string) *QueryBuilder {
	return &QueryBuilder{
		bucket:  bucket,
		filters: make(map[string]string),
		start:   time.Now().Add(-1 * time.Hour), // Default: last hour
		stop:    time.Now(),
	}
}

// From sets the measurement.
func (qb *QueryBuilder) From(measurement string) *QueryBuilder {
	qb.measurement = measurement
	return qb
}

// Range sets the time range.
func (qb *QueryBuilder) Range(start, stop time.Time) *QueryBuilder {
	qb.start = start
	qb.stop = stop
	return qb
}

// Filter adds a tag filter.
func (qb *QueryBuilder) Filter(tag, value string) *QueryBuilder {
	qb.filters[tag] = value
	return qb
}

// Aggregate adds an aggregation function.
func (qb *QueryBuilder) Aggregate(fn, window string) *QueryBuilder {
	qb.aggregation = fn
	qb.window = window
	return qb
}

// Build builds the Flux query.
func (qb *QueryBuilder) Build() string {
	query := fmt.Sprintf(`from(bucket: "%s")`, qb.bucket)
	query += fmt.Sprintf(` |> range(start: %s, stop: %s)`, qb.start.Format(time.RFC3339), qb.stop.Format(time.RFC3339))

	if qb.measurement != "" {
		query += fmt.Sprintf(` |> filter(fn: (r) => r._measurement == "%s")`, qb.measurement)
	}

	for tag, value := range qb.filters {
		query += fmt.Sprintf(` |> filter(fn: (r) => r.%s == "%s")`, tag, value)
	}

	if qb.aggregation != "" && qb.window != "" {
		query += fmt.Sprintf(` |> aggregateWindow(every: %s, fn: %s)`, qb.window, qb.aggregation)
	}

	return query
}
