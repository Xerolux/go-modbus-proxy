package manager

import (
	"fmt"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/proxy"
	"sync"
	"time"
)

// Manager manages proxies.
type Manager struct {
	mu      sync.RWMutex
	proxies map[string]*proxy.ProxyInstance
	cfgMgr  *config.Manager
	log     *logger.Logger
}

// NewManager creates a manager.
func NewManager(cfgMgr *config.Manager, log *logger.Logger) *Manager {
	m := &Manager{
		proxies: make(map[string]*proxy.ProxyInstance),
		cfgMgr:  cfgMgr,
		log:     log,
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

	p := proxy.NewProxyInstance(cfg.ID, cfg.Name, cfg.ListenAddr, cfg.TargetAddr, m.log)
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

// GetProxies returns status of all proxies.
func (m *Manager) GetProxies() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := []map[string]interface{}{}
	for _, p := range m.proxies {
		status := p.Stats
		uptime := time.Duration(0)
		if status.Status == "Running" {
			uptime = time.Since(status.LastStart)
		}

		res = append(res, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"listen_addr": p.ListenAddr,
			"target_addr": p.TargetAddr,
			"status":      status.Status,
			"uptime_s":    uptime.Seconds(),
			"requests":    status.Requests,
			"errors":      status.Errors,
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
