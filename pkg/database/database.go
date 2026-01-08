package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection and initializes the schema.
func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the database tables if they don't exist.
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		ip TEXT PRIMARY KEY,
		mac TEXT,
		name TEXT,
		first_seen DATETIME NOT NULL,
		last_connect DATETIME NOT NULL,
		request_count INTEGER DEFAULT 0,
		proxy_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS connection_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_ip TEXT NOT NULL,
		proxy_id TEXT NOT NULL,
		connected_at DATETIME NOT NULL,
		request_count INTEGER DEFAULT 1,
		FOREIGN KEY (device_ip) REFERENCES devices(ip) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_connection_history_device ON connection_history(device_ip);
	CREATE INDEX IF NOT EXISTS idx_connection_history_time ON connection_history(connected_at);
	CREATE INDEX IF NOT EXISTS idx_devices_proxy ON devices(proxy_id);

	-- Trigger to update updated_at on devices table
	CREATE TRIGGER IF NOT EXISTS update_devices_timestamp
	AFTER UPDATE ON devices
	BEGIN
		UPDATE devices SET updated_at = CURRENT_TIMESTAMP WHERE ip = NEW.ip;
	END;
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Device represents a connected Modbus device.
type Device struct {
	IP           string
	MAC          string
	Name         string
	FirstSeen    time.Time
	LastConnect  time.Time
	RequestCount int64
	ProxyID      string
}

// SaveDevice saves or updates a device in the database.
func (db *DB) SaveDevice(device *Device) error {
	query := `
		INSERT INTO devices (ip, mac, name, first_seen, last_connect, request_count, proxy_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET
			mac = excluded.mac,
			name = CASE WHEN excluded.name != '' THEN excluded.name ELSE devices.name END,
			last_connect = excluded.last_connect,
			request_count = excluded.request_count,
			proxy_id = excluded.proxy_id
	`

	_, err := db.conn.Exec(query,
		device.IP,
		device.MAC,
		device.Name,
		device.FirstSeen,
		device.LastConnect,
		device.RequestCount,
		device.ProxyID,
	)

	return err
}

// GetDevice retrieves a device by IP address.
func (db *DB) GetDevice(ip string) (*Device, error) {
	query := `SELECT ip, mac, name, first_seen, last_connect, request_count, proxy_id FROM devices WHERE ip = ?`

	var device Device
	err := db.conn.QueryRow(query, ip).Scan(
		&device.IP,
		&device.MAC,
		&device.Name,
		&device.FirstSeen,
		&device.LastConnect,
		&device.RequestCount,
		&device.ProxyID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &device, nil
}

// GetAllDevices retrieves all devices from the database.
func (db *DB) GetAllDevices() ([]*Device, error) {
	query := `SELECT ip, mac, name, first_seen, last_connect, request_count, proxy_id FROM devices ORDER BY last_connect DESC`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		var device Device
		err := rows.Scan(
			&device.IP,
			&device.MAC,
			&device.Name,
			&device.FirstSeen,
			&device.LastConnect,
			&device.RequestCount,
			&device.ProxyID,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}

	return devices, rows.Err()
}

// UpdateDeviceName updates only the name of a device.
func (db *DB) UpdateDeviceName(ip, name string) error {
	query := `UPDATE devices SET name = ? WHERE ip = ?`
	_, err := db.conn.Exec(query, name, ip)
	return err
}

// IncrementRequestCount increments the request counter for a device.
func (db *DB) IncrementRequestCount(ip string) error {
	query := `UPDATE devices SET request_count = request_count + 1, last_connect = ? WHERE ip = ?`
	_, err := db.conn.Exec(query, time.Now(), ip)
	return err
}

// ConnectionHistoryEntry represents a single connection history record.
type ConnectionHistoryEntry struct {
	ID           int64
	DeviceIP     string
	ProxyID      string
	ConnectedAt  time.Time
	RequestCount int
}

// AddConnectionHistory adds a new connection history entry.
func (db *DB) AddConnectionHistory(deviceIP, proxyID string) error {
	query := `INSERT INTO connection_history (device_ip, proxy_id, connected_at) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, deviceIP, proxyID, time.Now())
	return err
}

// GetConnectionHistory retrieves connection history for a device.
func (db *DB) GetConnectionHistory(deviceIP string, limit int) ([]*ConnectionHistoryEntry, error) {
	query := `
		SELECT id, device_ip, proxy_id, connected_at, request_count
		FROM connection_history
		WHERE device_ip = ?
		ORDER BY connected_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, deviceIP, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*ConnectionHistoryEntry
	for rows.Next() {
		var entry ConnectionHistoryEntry
		err := rows.Scan(
			&entry.ID,
			&entry.DeviceIP,
			&entry.ProxyID,
			&entry.ConnectedAt,
			&entry.RequestCount,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// GetAllConnectionHistory retrieves all connection history with optional filters.
func (db *DB) GetAllConnectionHistory(proxyID string, limit int) ([]*ConnectionHistoryEntry, error) {
	var query string
	var args []interface{}

	if proxyID != "" {
		query = `
			SELECT id, device_ip, proxy_id, connected_at, request_count
			FROM connection_history
			WHERE proxy_id = ?
			ORDER BY connected_at DESC
			LIMIT ?
		`
		args = []interface{}{proxyID, limit}
	} else {
		query = `
			SELECT id, device_ip, proxy_id, connected_at, request_count
			FROM connection_history
			ORDER BY connected_at DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*ConnectionHistoryEntry
	for rows.Next() {
		var entry ConnectionHistoryEntry
		err := rows.Scan(
			&entry.ID,
			&entry.DeviceIP,
			&entry.ProxyID,
			&entry.ConnectedAt,
			&entry.RequestCount,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// CleanOldHistory removes connection history older than the specified duration.
func (db *DB) CleanOldHistory(olderThan time.Duration) error {
	query := `DELETE FROM connection_history WHERE connected_at < ?`
	cutoff := time.Now().Add(-olderThan)
	_, err := db.conn.Exec(query, cutoff)
	return err
}

// GetDeviceStats returns statistics about devices.
func (db *DB) GetDeviceStats() (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_devices,
			SUM(request_count) as total_requests,
			COUNT(DISTINCT proxy_id) as unique_proxies
		FROM devices
	`

	var totalDevices, totalRequests, uniqueProxies int64
	err := db.conn.QueryRow(query).Scan(&totalDevices, &totalRequests, &uniqueProxies)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_devices":   totalDevices,
		"total_requests":  totalRequests,
		"unique_proxies":  uniqueProxies,
	}, nil
}
