package config

import (
	"encoding/json"
	"os"
	"sync"
)

// ProxyConfig defines the configuration for a single proxy instance.
type ProxyConfig struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ListenAddr        string `json:"listen_addr"`
	TargetAddr        string `json:"target_addr"`
	Enabled           bool   `json:"enabled"`
	Paused            bool   `json:"paused"`             // New: Paused state (different from Enabled)
	ConnectionTimeout int    `json:"connection_timeout"` // New: Connection timeout in seconds (default: 10)
	ReadTimeout       int    `json:"read_timeout"`       // New: Read timeout in seconds (default: 30)
	MaxRetries        int    `json:"max_retries"`        // New: Max retry attempts (default: 3)
	Description       string `json:"description"`        // New: User description
}

// Config holds the global configuration.
type Config struct {
	WebPort       string        `json:"web_port"`
	AdminPassHash string        `json:"admin_pass_hash"` // Empty means first-run
	Proxies       []ProxyConfig `json:"proxies"`
}

// Manager handles config persistence.
type Manager struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

// NewManager creates a config manager.
func NewManager(path string) *Manager {
	return &Manager{
		path: path,
		cfg: Config{
			WebPort: ":8080",
			Proxies: []ProxyConfig{},
		},
	}
}

// Load reads config from disk.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Open(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&m.cfg)
}

// Save writes config to disk.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Create(m.path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m.cfg)
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Deep copy if needed, but for now value return is fine for struct (slices share backing array though)
	// To be safe, we should copy the slice.
	c := m.cfg
	if c.Proxies != nil {
		proxies := make([]ProxyConfig, len(c.Proxies))
		copy(proxies, c.Proxies)
		c.Proxies = proxies
	}
	return c
}

func (m *Manager) Update(fn func(*Config) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to modify
	newCfg := m.cfg
	// (Deep copy slice for safety)
	if newCfg.Proxies != nil {
		proxies := make([]ProxyConfig, len(newCfg.Proxies))
		copy(proxies, newCfg.Proxies)
		newCfg.Proxies = proxies
	}

	if err := fn(&newCfg); err != nil {
		return err
	}

	m.cfg = newCfg

	// Save to disk immediately
	f, err := os.Create(m.path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m.cfg)
}
