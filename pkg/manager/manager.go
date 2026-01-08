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
}

// NewManager creates a manager with database support.
func NewManager(cfgMgr *config.Manager, log *logger.Logger, db *database.DB) *Manager {
	m := &Manager{
		proxies:       make(map[string]*proxy.ProxyInstance),
		cfgMgr:        cfgMgr,
		log:           log,
		deviceTracker: devices.NewTracker(db),
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

	p := proxy.NewProxyInstance(cfg.ID, cfg.Name, cfg.ListenAddr, cfg.TargetAddr, m.log, m.deviceTracker)
	m.proxies[cfg.ID] = p

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

	if p, ok := m.proxies[id]; ok {
		p.Stop()
		delete(m.proxies, id)
	}

	return m.cfgMgr.Update(func(c *config.Config) error {
		newProxies := []config.ProxyConfig{}
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
	p := proxy.NewProxyInstance(cfg.ID, cfg.Name, cfg.ListenAddr, cfg.TargetAddr, m.log, m.deviceTracker)
	m.proxies[cfg.ID] = p

	// Start if it was enabled and not paused
	if cfg.Enabled && !cfg.Paused {
		p.Start()
	}

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

	res := []map[string]interface{}{}
	for _, p := range m.proxies {
		status := p.Stats
		uptime := time.Duration(0)
		if status.Status == "Running" {
			uptime = time.Since(status.LastStart)
		}

		// Get config for this proxy
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
			"requests":           status.Requests,
			"errors":             status.Errors,
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
