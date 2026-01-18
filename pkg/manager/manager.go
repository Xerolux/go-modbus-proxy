package manager

import (
	"fmt"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/database"
	"modbusproxy/pkg/devices"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/proxy"
	"sync"
	"time"
)

// Manager manages proxies.
type Manager struct {
	mu            sync.RWMutex
	proxies       map[string]*proxy.ProxyInstance
	cfgMgr        *config.Manager
	log           *logger.Logger
	deviceTracker *devices.Tracker
	broadcaster   *EventBroadcaster
}

// NewManager creates a manager with database support.
func NewManager(cfgMgr *config.Manager, log *logger.Logger, db *database.DB) *Manager {
	m := &Manager{
		proxies:       make(map[string]*proxy.ProxyInstance),
		cfgMgr:        cfgMgr,
		log:           log,
		deviceTracker: devices.NewTracker(db),
		broadcaster:   NewEventBroadcaster(),
	}
	return m
}

// Initialize loads config and starts enabled proxies.
func (m *Manager) Initialize() {
	cfg := m.cfgMgr.Get()
	for _, pCfg := range cfg.Proxies {
		m.AddProxy(pCfg, false) // Add but don't save to config again
		if pCfg.Enabled {
			m.StartProxy(pCfg.ID)
		}
	}
}

// AddProxy adds a new proxy or updates existing.
func (m *Manager) AddProxy(cfg config.ProxyConfig, save bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing if any
	if old, ok := m.proxies[cfg.ID]; ok {
		old.Stop()
	}

	p := proxy.NewProxyInstance(cfg.ID, cfg.Name, cfg.ListenAddr, cfg.TargetAddr, cfg.MaxReadSize, m.log, m.deviceTracker)
	m.proxies[cfg.ID] = p

	// Broadcast event
	m.broadcaster.Broadcast(map[string]interface{}{
		"type":      "proxy_added",
		"timestamp": time.Now(),
		"proxy":     m.getProxyStatus(cfg.ID),
	})

	if save {
		return m.cfgMgr.Update(func(c *config.Config) error {
			// Check if exists
			found := false
			for i, pc := range c.Proxies {
				if pc.ID == cfg.ID {
					c.Proxies[i] = cfg
					found = true
					break
				}
			}
			if !found {
				c.Proxies = append(c.Proxies, cfg)
			}
			return nil
		})
	}
	return nil
}

// RemoveProxy removes a proxy.
func (m *Manager) RemoveProxy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.proxies[id]; !ok {
		return fmt.Errorf("proxy not found")
	}

	p := m.proxies[id]
	p.Stop()
	delete(m.proxies, id)

	// Broadcast event
	m.broadcaster.Broadcast(map[string]interface{}{
		"type":      "proxy_removed",
		"timestamp": time.Now(),
		"proxy_id":  id,
	})

	return m.cfgMgr.Update(func(c *config.Config) error {
		newProxies := make([]config.ProxyConfig, 0, len(c.Proxies))
		for _, pc := range c.Proxies {
			if pc.ID != id {
				newProxies = append(newProxies, pc)
			}
		}
		c.Proxies = newProxies
		return nil
	})
}

// StartProxy starts a proxy.
func (m *Manager) StartProxy(id string) error {
	m.mu.Lock()
	p, ok := m.proxies[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("proxy not found")
	}

	if err := p.Start(); err != nil {
		return err
	}

	// Broadcast event
	m.broadcaster.Broadcast(map[string]interface{}{
		"type":      "proxy_started",
		"timestamp": time.Now(),
		"proxy":     m.getProxyStatus(id),
	})

	// Update enabled state
	return m.cfgMgr.Update(func(c *config.Config) error {
		for i, pc := range c.Proxies {
			if pc.ID == id {
				c.Proxies[i].Enabled = true
			}
		}
		return nil
	})
}

// StopProxy stops a proxy.
func (m *Manager) StopProxy(id string) error {
	m.mu.Lock()
	p, ok := m.proxies[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("proxy not found")
	}

	p.Stop()

	// Broadcast event
	m.broadcaster.Broadcast(map[string]interface{}{
		"type":      "proxy_stopped",
		"timestamp": time.Now(),
		"proxy":     m.getProxyStatus(id),
	})

	// Update enabled state
	return m.cfgMgr.Update(func(c *config.Config) error {
		for i, pc := range c.Proxies {
			if pc.ID == id {
				c.Proxies[i].Enabled = false
			}
		}
		return nil
	})
}

// PauseProxy pauses a running proxy.
func (m *Manager) PauseProxy(id string) error {
	m.mu.Lock()
	p, ok := m.proxies[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("proxy not found")
	}

	// Stop the proxy but keep enabled=true, set paused=true
	p.Stop()

	return m.cfgMgr.Update(func(c *config.Config) error {
		for i, pc := range c.Proxies {
			if pc.ID == id {
				c.Proxies[i].Paused = true
			}
		}
		return nil
	})
}

// ResumeProxy resumes a paused proxy.
func (m *Manager) ResumeProxy(id string) error {
	m.mu.Lock()
	p, ok := m.proxies[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("proxy not found")
	}

	// Start the proxy and set paused=false
	if err := p.Start(); err != nil {
		return err
	}

	return m.cfgMgr.Update(func(c *config.Config) error {
		for i, pc := range c.Proxies {
			if pc.ID == id {
				c.Proxies[i].Paused = false
				c.Proxies[i].Enabled = true
			}
		}
		return nil
	})
}

// UpdateProxy updates an existing proxy configuration.
func (m *Manager) UpdateProxy(cfg config.ProxyConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if proxy exists
	if _, ok := m.proxies[cfg.ID]; !ok {
		return fmt.Errorf("proxy not found")
	}

	// Stop the old proxy
	m.proxies[cfg.ID].Stop()

	// Create new proxy with updated config
	p := proxy.NewProxyInstance(cfg.ID, cfg.Name, cfg.ListenAddr, cfg.TargetAddr, cfg.MaxReadSize, m.log, m.deviceTracker)
	m.proxies[cfg.ID] = p

	// Start if it was enabled and not paused
	if cfg.Enabled && !cfg.Paused {
		p.Start()
	}

	// Broadcast event
	m.broadcaster.Broadcast(map[string]interface{}{
		"type":      "proxy_updated",
		"timestamp": time.Now(),
		"proxy":     m.getProxyStatus(cfg.ID),
	})

	// Update config
	return m.cfgMgr.Update(func(c *config.Config) error {
		for i, pc := range c.Proxies {
			if pc.ID == cfg.ID {
				c.Proxies[i] = cfg
				break
			}
		}
		return nil
	})
}

// GetProxies returns status of all proxies.
func (m *Manager) GetProxies() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get current config to include paused status and other fields
	cfg := m.cfgMgr.Get()
	cfgMap := make(map[string]config.ProxyConfig)
	for _, pc := range cfg.Proxies {
		cfgMap[pc.ID] = pc
	}

	res := make([]map[string]interface{}, 0, len(m.proxies))
	for _, p := range m.proxies {
		status := &p.Stats
		uptime := time.Duration(0)
		if status.Status == "Running" {
			uptime = time.Since(status.LastStart)
		}

		pCfg := cfgMap[p.ID]

		res = append(res, map[string]interface{}{
			"id":                 p.ID,
			"name":               p.Name,
			"listen_addr":        p.ListenAddr,
			"target_addr":        p.TargetAddr,
			"status":             status.Status,
			"paused":             pCfg.Paused,
			"enabled":            pCfg.Enabled,
			"uptime_s":           uptime.Seconds(),
			"requests":           status.Requests.Load(),
			"errors":             status.Errors.Load(),
			"description":        pCfg.Description,
			"connection_timeout": pCfg.ConnectionTimeout,
			"read_timeout":       pCfg.ReadTimeout,
			"max_retries":        pCfg.MaxRetries,
		})
	}
	return res
}

// StopAll stops all running proxies.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.proxies {
		if p.Stats.Status == "Running" {
			p.Stop()
		}
	}
}

// StartAll starts all enabled proxies.
func (m *Manager) StartAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := m.cfgMgr.Get()
	cfgMap := make(map[string]config.ProxyConfig)
	for _, pc := range cfg.Proxies {
		cfgMap[pc.ID] = pc
	}

	for _, p := range m.proxies {
		pCfg := cfgMap[p.ID]
		if pCfg.Enabled && !pCfg.Paused && p.Stats.Status != "Running" {
			if err := p.Start(); err != nil {
				m.log.Error(p.ID, fmt.Sprintf("Failed to start: %v", err))
			}
		}
	}
}

// GetDevices returns all tracked devices.
func (m *Manager) GetDevices() []devices.Device {
	return m.deviceTracker.GetDevices()
}

// SetDeviceName sets a user-friendly name for a device.
func (m *Manager) SetDeviceName(ip, name string) error {
	return m.deviceTracker.SetDeviceName(ip, name)
}

// GetConnectionHistory returns connection history for a device.
func (m *Manager) GetConnectionHistory(ip string, limit int) ([]*database.ConnectionHistoryEntry, error) {
	return m.deviceTracker.GetConnectionHistory(ip, limit)
}

// GetAllConnectionHistory returns all connection history with optional proxy filter.
func (m *Manager) GetAllConnectionHistory(proxyID string, limit int) ([]*database.ConnectionHistoryEntry, error) {
	return m.deviceTracker.GetAllConnectionHistory(proxyID, limit)
}

// GetProxyEventsSubscription returns a subscription for proxy events
func (m *Manager) GetProxyEventsSubscription() chan interface{} {
	return m.broadcaster.Subscribe()
}

// UnsubscribeProxyEvents unsubscribes from proxy events
func (m *Manager) UnsubscribeProxyEvents(ch chan interface{}) {
	m.broadcaster.Unsubscribe(ch)
}

// getProxyStatus returns proxy status information
func (m *Manager) getProxyStatus(id string) map[string]interface{} {
	p, ok := m.proxies[id]
	if !ok {
		return nil
	}

	cfg := m.cfgMgr.Get()
	cfgMap := make(map[string]config.ProxyConfig)
	if cfg.Proxies != nil {
		for _, pc := range cfg.Proxies {
			cfgMap[pc.ID] = pc
		}
	}

	status := &p.Stats
	uptime := time.Duration(0)
	if status.Status == "Running" {
		uptime = time.Since(status.LastStart)
	}

	pCfg := cfgMap[p.ID]

	return map[string]interface{}{
		"id":                 p.ID,
		"name":               p.Name,
		"listen_addr":        p.ListenAddr,
		"target_addr":        p.TargetAddr,
		"status":             status.Status,
		"paused":             pCfg.Paused,
		"enabled":            pCfg.Enabled,
		"uptime_s":           uptime.Seconds(),
		"requests":           status.Requests.Load(),
		"errors":             status.Errors.Load(),
		"description":        pCfg.Description,
		"connection_timeout": pCfg.ConnectionTimeout,
		"read_timeout":       pCfg.ReadTimeout,
		"max_retries":        pCfg.MaxRetries,
	}
}
