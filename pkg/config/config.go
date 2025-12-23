package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Config holds the proxy configuration.
type Config struct {
	ListenAddr string `json:"listen_addr"`
	TargetAddr string `json:"target_addr"`
	WebAddr    string `json:"web_addr"`
}

// Manager manages the configuration safely.
type Manager struct {
	mu     sync.RWMutex
	config Config
	path   string
}

// NewManager creates a new configuration manager.
func NewManager(path string) *Manager {
	return &Manager{
		path: path,
		config: Config{
			ListenAddr: ":5020",
			TargetAddr: "127.0.0.1:502",
			WebAddr:    ":8080",
		},
	}
}

// Load loads the configuration from the file.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Open(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Use defaults
		}
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&m.config)
}

// Save saves the current configuration to the file.
func (m *Manager) Save() error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	f, err := os.Create(m.path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// Get returns the current configuration.
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Update updates the configuration and saves it.
func (m *Manager) Update(newConfig Config) error {
	m.mu.Lock()
	m.config = newConfig
	m.mu.Unlock()
	return m.Save()
}
