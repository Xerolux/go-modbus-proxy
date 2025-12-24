package config

import (
	"fmt"
	"os"
	"strings"
)

// ExportEnv exports configuration as environment variables.
// This is useful for containerized deployments and 12-factor apps.
func ExportEnv(cfg *Config) {
	// Server configuration
	if cfg.WebPort != "" {
		os.Setenv("MODBRIDGE_WEB_PORT", cfg.WebPort)
	}

	// Export proxy count
	os.Setenv("MODBRIDGE_PROXY_COUNT", fmt.Sprintf("%d", len(cfg.Proxies)))

	// Export individual proxy configs
	for i, proxy := range cfg.Proxies {
		prefix := fmt.Sprintf("MODBRIDGE_PROXY_%d", i)
		os.Setenv(prefix+"_ID", proxy.ID)
		os.Setenv(prefix+"_NAME", proxy.Name)
		os.Setenv(prefix+"_LISTEN", proxy.ListenAddr)
		os.Setenv(prefix+"_TARGET", proxy.TargetAddr)
		os.Setenv(prefix+"_POOL_SIZE", fmt.Sprintf("%d", proxy.PoolSize))
		os.Setenv(prefix+"_ENABLED", fmt.Sprintf("%t", proxy.Enabled))
	}
}

// GenerateEnvFile generates a .env file from configuration.
func GenerateEnvFile(cfg *Config, filename string) error {
	var sb strings.Builder

	sb.WriteString("# Modbridge Configuration\n")
	sb.WriteString("# Generated automatically from config.json\n\n")

	// Server configuration
	sb.WriteString("# Server Configuration\n")
	sb.WriteString(fmt.Sprintf("MODBRIDGE_WEB_PORT=%s\n", cfg.WebPort))
	sb.WriteString("\n")

	// Proxies
	sb.WriteString("# Proxy Configuration\n")
	sb.WriteString(fmt.Sprintf("MODBRIDGE_PROXY_COUNT=%d\n\n", len(cfg.Proxies)))

	for i, proxy := range cfg.Proxies {
		sb.WriteString(fmt.Sprintf("# Proxy %d: %s\n", i, proxy.Name))
		prefix := fmt.Sprintf("MODBRIDGE_PROXY_%d", i)
		sb.WriteString(fmt.Sprintf("%s_ID=%s\n", prefix, proxy.ID))
		sb.WriteString(fmt.Sprintf("%s_NAME=%s\n", prefix, proxy.Name))
		sb.WriteString(fmt.Sprintf("%s_LISTEN=%s\n", prefix, proxy.ListenAddr))
		sb.WriteString(fmt.Sprintf("%s_TARGET=%s\n", prefix, proxy.TargetAddr))
		sb.WriteString(fmt.Sprintf("%s_POOL_SIZE=%d\n", prefix, proxy.PoolSize))
		sb.WriteString(fmt.Sprintf("%s_ENABLED=%t\n", prefix, proxy.Enabled))
		sb.WriteString("\n")
	}

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}

// LoadFromEnv loads configuration from environment variables.
// This provides an alternative to JSON configuration files.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		WebPort: getEnv("MODBRIDGE_WEB_PORT", ":8080"),
		Proxies: make([]ProxyConfig, 0),
	}

	// Load proxies
	proxyCount := getEnvInt("MODBRIDGE_PROXY_COUNT", 0)
	for i := 0; i < proxyCount; i++ {
		prefix := fmt.Sprintf("MODBRIDGE_PROXY_%d", i)
		proxy := ProxyConfig{
			ID:         getEnv(prefix+"_ID", ""),
			Name:       getEnv(prefix+"_NAME", ""),
			ListenAddr: getEnv(prefix+"_LISTEN", ""),
			TargetAddr: getEnv(prefix+"_TARGET", ""),
			PoolSize:   getEnvInt(prefix+"_POOL_SIZE", 10),
			Enabled:    getEnvBool(prefix+"_ENABLED", true),
		}

		if proxy.ID != "" && proxy.ListenAddr != "" && proxy.TargetAddr != "" {
			cfg.Proxies = append(cfg.Proxies, proxy)
		}
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt gets an integer environment variable with a default value.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return defaultValue
	}

	return result
}

// getEnvBool gets a boolean environment variable with a default value.
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	value = strings.ToLower(value)
	return value == "true" || value == "1" || value == "yes" || value == "on"
}
