package devices

import (
	"modbusproxy/pkg/database"
	"net"
	"sync"
	"time"
)

// Device represents a connected client device.
type Device struct {
	IP           string    `json:"ip"`
	MAC          string    `json:"mac"`
	Name         string    `json:"name"` // User-friendly name
	LastConnect  time.Time `json:"last_connect"`
	FirstSeen    time.Time `json:"first_seen"`
	RequestCount int64     `json:"request_count"`
	ProxyID      string    `json:"proxy_id"`
}

// Tracker tracks connected devices with persistent storage.
type Tracker struct {
	mu      sync.RWMutex
	devices map[string]*Device // key: IP address (cache)
	db      *database.DB       // persistent storage
}

// NewTracker creates a new device tracker with database support.
func NewTracker(db *database.DB) *Tracker {
	t := &Tracker{
		devices: make(map[string]*Device),
		db:      db,
	}

	// Load existing devices from database into cache
	if db != nil {
		t.loadDevicesFromDB()
	}

	return t
}

// loadDevicesFromDB loads all devices from the database into the cache.
func (t *Tracker) loadDevicesFromDB() {
	dbDevices, err := t.db.GetAllDevices()
	if err != nil {
		// Log error but don't fail - can continue with empty cache
		return
	}

	for _, dbDev := range dbDevices {
		t.devices[dbDev.IP] = &Device{
			IP:           dbDev.IP,
			MAC:          dbDev.MAC,
			Name:         dbDev.Name,
			FirstSeen:    dbDev.FirstSeen,
			LastConnect:  dbDev.LastConnect,
			RequestCount: dbDev.RequestCount,
			ProxyID:      dbDev.ProxyID,
		}
	}
}

// TrackConnection records a connection from a client.
func (t *Tracker) TrackConnection(conn net.Conn, proxyID string) {
	if conn == nil {
		return
	}

	remoteAddr := conn.RemoteAddr()
	if remoteAddr == nil {
		return
	}

	ip, _, err := net.SplitHostPort(remoteAddr.String())
	if err != nil {
		ip = remoteAddr.String()
	}

	mac := getMACAddress(ip)
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	device, exists := t.devices[ip]
	if !exists {
		device = &Device{
			IP:        ip,
			MAC:       mac,
			Name:      "", // Will be set by user
			FirstSeen: now,
			ProxyID:   proxyID,
		}
		t.devices[ip] = device
	}

	device.LastConnect = now
	device.RequestCount++
	device.ProxyID = proxyID

	// Save to database if available
	if t.db != nil {
		dbDevice := &database.Device{
			IP:           device.IP,
			MAC:          device.MAC,
			Name:         device.Name,
			FirstSeen:    device.FirstSeen,
			LastConnect:  device.LastConnect,
			RequestCount: device.RequestCount,
			ProxyID:      device.ProxyID,
		}
		// Fire and forget - don't block on DB writes
		go func() {
			_ = t.db.SaveDevice(dbDevice)
			// Also add to connection history
			_ = t.db.AddConnectionHistory(ip, proxyID)
		}()
	}
}

// GetDevices returns all tracked devices.
func (t *Tracker) GetDevices() []Device {
	t.mu.RLock()
	defer t.mu.RUnlock()

	devices := make([]Device, 0, len(t.devices))
	for _, d := range t.devices {
		devices = append(devices, *d)
	}
	return devices
}

// SetDeviceName sets a user-friendly name for a device.
func (t *Tracker) SetDeviceName(ip, name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	device, exists := t.devices[ip]
	if !exists {
		// Create device entry if it doesn't exist
		device = &Device{
			IP:        ip,
			Name:      name,
			FirstSeen: time.Now(),
		}
		t.devices[ip] = device
	} else {
		device.Name = name
	}

	// Save to database if available
	if t.db != nil {
		go func() {
			_ = t.db.UpdateDeviceName(ip, name)
		}()
	}

	return nil
}

// GetConnectionHistory returns connection history for a device.
func (t *Tracker) GetConnectionHistory(ip string, limit int) ([]*database.ConnectionHistoryEntry, error) {
	if t.db == nil {
		return nil, nil
	}
	return t.db.GetConnectionHistory(ip, limit)
}

// GetAllConnectionHistory returns all connection history with optional proxy filter.
func (t *Tracker) GetAllConnectionHistory(proxyID string, limit int) ([]*database.ConnectionHistoryEntry, error) {
	if t.db == nil {
		return nil, nil
	}
	return t.db.GetAllConnectionHistory(proxyID, limit)
}

// getMACAddress attempts to get the MAC address for an IP.
// Note: This is limited and may not work in all scenarios.
func getMACAddress(ip string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	// Check if it's localhost
	if parsedIP.IsLoopback() {
		return "localhost"
	}

	// Try to find MAC from ARP (limited, mainly works for local network)
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var netIP net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				netIP = v.IP
			case *net.IPAddr:
				netIP = v.IP
			}

			if netIP != nil && netIP.Equal(parsedIP) {
				return iface.HardwareAddr.String()
			}
		}
	}

	// If we can't find it, return unknown
	// In production, you might want to use ARP table parsing or
	// external tools for more accurate MAC resolution
	return "unknown"
}
