package config

import (
	"encoding/json"
	"os"
	"sync"
)

// ProxyConfig defines the configuration for a single proxy instance.
type ProxyConfig struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	ListenAddr        string   `json:"listen_addr"`
	TargetAddr        string   `json:"target_addr"`
	Enabled           bool     `json:"enabled"`
	Paused            bool     `json:"paused"`             // New: Paused state (different from Enabled)
	ConnectionTimeout int      `json:"connection_timeout"` // New: Connection timeout in seconds (default: 10)
	ReadTimeout       int      `json:"read_timeout"`       // New: Read timeout in seconds (default: 30)
	MaxRetries        int      `json:"max_retries"`        // New: Max retry attempts (default: 3)
	Description       string   `json:"description"`        // New: User description
	MaxReadSize       int      `json:"max_read_size"`
	Tags              []string `json:"tags"`
}

// Config holds the global configuration.
type Config struct {
	WebPort             string        `json:"web_port"`
	AdminPassHash       string        `json:"admin_pass_hash"`       // Empty means first-run
	ForcePasswordChange bool          `json:"force_password_change"` // Forces password change on next login
	Proxies             []ProxyConfig `json:"proxies"`

	LogLevel      string `json:"log_level"`
	LogMaxSize    int    `json:"log_max_size"`
	LogMaxFiles   int    `json:"log_max_files"`
	LogMaxAgeDays int    `json:"log_max_age_days"`

	TLSEnabled     bool   `json:"tls_enabled"`
	TLSCertFile    string `json:"tls_cert_file"`
	TLSKeyFile     string `json:"tls_key_file"`
	SessionTimeout int    `json:"session_timeout"`

	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`

	RateLimitEnabled  bool `json:"rate_limit_enabled"`
	RateLimitRequests int  `json:"rate_limit_requests"`
	RateLimitBurst    int  `json:"rate_limit_burst"`

	IPWhitelistEnabled bool     `json:"ip_whitelist_enabled"`
	IPWhitelist        []string `json:"ip_whitelist"`
	IPBlacklistEnabled bool     `json:"ip_blacklist_enabled"`
	IPBlacklist        []string `json:"ip_blacklist"`

	EmailEnabled        bool   `json:"email_enabled"`
	EmailSMTPServer     string `json:"email_smtp_server"`
	EmailSMTPPort       int    `json:"email_smtp_port"`
	EmailFrom           string `json:"email_from"`
	EmailTo             string `json:"email_to"`
	EmailUsername       string `json:"email_username"`
	EmailPassword       string `json:"email_password"`
	EmailAlertOnError   bool   `json:"email_alert_on_error"`
	EmailAlertOnWarning bool   `json:"email_alert_on_warning"`

	BackupEnabled   bool   `json:"backup_enabled"`
	BackupInterval  string `json:"backup_interval"`
	BackupRetention int    `json:"backup_retention"`
	BackupPath      string `json:"backup_path"`
	BackupDatabase  bool   `json:"backup_database"`
	BackupConfig    bool   `json:"backup_config"`

	MetricsEnabled bool   `json:"metrics_enabled"`
	MetricsPort    string `json:"metrics_port"`

	DebugMode      bool `json:"debug_mode"`
	MaxConnections int  `json:"max_connections"`
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
			WebPort:             ":8080",
			Proxies:             []ProxyConfig{},
			LogLevel:            "INFO",
			LogMaxSize:          100,
			LogMaxFiles:         10,
			LogMaxAgeDays:       30,
			TLSEnabled:          false,
			SessionTimeout:      24,
			CORSAllowedOrigins:  []string{"*"},
			CORSAllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			CORSAllowedHeaders:  []string{"Content-Type", "Authorization"},
			RateLimitEnabled:    true,
			RateLimitRequests:   60,
			RateLimitBurst:      100,
			IPWhitelistEnabled:  false,
			IPBlacklistEnabled:  false,
			EmailEnabled:        false,
			EmailAlertOnError:   true,
			EmailAlertOnWarning: false,
			BackupEnabled:       true,
			BackupInterval:      "daily",
			BackupRetention:     7,
			BackupPath:          "./backups",
			BackupDatabase:      true,
			BackupConfig:        true,
			MetricsEnabled:      true,
			MetricsPort:         ":9090",
			DebugMode:           false,
			MaxConnections:      1000,
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

	c := m.cfg
	if c.Proxies != nil {
		proxies := make([]ProxyConfig, len(c.Proxies))
		copy(proxies, c.Proxies)
		c.Proxies = proxies
	}
	if c.CORSAllowedOrigins != nil {
		origins := make([]string, len(c.CORSAllowedOrigins))
		copy(origins, c.CORSAllowedOrigins)
		c.CORSAllowedOrigins = origins
	}
	if c.IPWhitelist != nil {
		whitelist := make([]string, len(c.IPWhitelist))
		copy(whitelist, c.IPWhitelist)
		c.IPWhitelist = whitelist
	}
	if c.IPBlacklist != nil {
		blacklist := make([]string, len(c.IPBlacklist))
		copy(blacklist, c.IPBlacklist)
		c.IPBlacklist = blacklist
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
	if newCfg.CORSAllowedOrigins != nil {
		origins := make([]string, len(newCfg.CORSAllowedOrigins))
		copy(origins, newCfg.CORSAllowedOrigins)
		newCfg.CORSAllowedOrigins = origins
	}
	if newCfg.IPWhitelist != nil {
		whitelist := make([]string, len(newCfg.IPWhitelist))
		copy(whitelist, newCfg.IPWhitelist)
		newCfg.IPWhitelist = whitelist
	}
	if newCfg.IPBlacklist != nil {
		blacklist := make([]string, len(newCfg.IPBlacklist))
		copy(blacklist, newCfg.IPBlacklist)
		newCfg.IPBlacklist = blacklist
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
