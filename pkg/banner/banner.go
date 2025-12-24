package banner

import (
	"fmt"
	"runtime"
	"time"
)

const (
	// Version is the application version
	Version = "1.0.0"

	// BuildDate is set during build
	BuildDate = "2024-01-15"

	// GitCommit is set during build
	GitCommit = "dev"
)

// Print displays the startup banner with configuration summary.
func Print(config BannerConfig) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                               ║")
	fmt.Println("║   ███╗   ███╗ ██████╗ ██████╗ ██████╗ ██████╗ ██╗██████╗ ██████╗ ███████╗")
	fmt.Println("║   ████╗ ████║██╔═══██╗██╔══██╗██╔══██╗██╔══██╗██║██╔══██╗██╔════╝ ██╔════╝")
	fmt.Println("║   ██╔████╔██║██║   ██║██║  ██║██████╔╝██████╔╝██║██║  ██║██║  ███╗█████╗  ")
	fmt.Println("║   ██║╚██╔╝██║██║   ██║██║  ██║██╔══██╗██╔══██╗██║██║  ██║██║   ██║██╔══╝  ")
	fmt.Println("║   ██║ ╚═╝ ██║╚██████╔╝██████╔╝██████╔╝██║  ██║██║██████╔╝╚██████╔╝███████╗")
	fmt.Println("║   ╚═╝     ╚═╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝╚═════╝  ╚═════╝ ╚══════╝")
	fmt.Println("║                                                               ║")
	fmt.Println("║           Enterprise-Grade Modbus TCP Proxy Manager          ║")
	fmt.Println("║                                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Version information
	fmt.Printf("  Version:       %s\n", Version)
	fmt.Printf("  Build Date:    %s\n", BuildDate)
	fmt.Printf("  Git Commit:    %s\n", GitCommit)
	fmt.Printf("  Go Version:    %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:       %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	// Configuration summary
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║ Configuration Summary                                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("  API Server:    %s\n", config.ServerAddr)
	if config.TLSEnabled {
		fmt.Println("  TLS:           Enabled ✓")
	} else {
		fmt.Println("  TLS:           Disabled")
	}

	fmt.Printf("  Database:      %s (%s)\n", config.DatabaseType, config.DatabasePath)
	fmt.Printf("  Log Level:     %s\n", config.LogLevel)
	fmt.Printf("  Proxies:       %d configured\n", config.ProxyCount)

	if config.ProxyCount > 0 {
		fmt.Println()
		fmt.Println("  Active Proxies:")
		for _, proxy := range config.Proxies {
			status := "disabled"
			if proxy.Enabled {
				status = "enabled"
			}
			fmt.Printf("    • %-20s  %s → %s  [%s]\n",
				proxy.Name, proxy.Listen, proxy.Target, status)
		}
	}

	fmt.Println()

	// Feature flags
	if len(config.Features) > 0 {
		fmt.Println("  Features:")
		for feature, enabled := range config.Features {
			status := "✓"
			if !enabled {
				status = "✗"
			}
			fmt.Printf("    %s %-30s\n", status, feature)
		}
		fmt.Println()
	}

	// Performance settings
	if config.ShowPerformance {
		fmt.Println("  Performance:")
		fmt.Printf("    Connection Pooling:   %s\n", boolToStatus(config.ConnectionPooling))
		fmt.Printf("    DNS Caching:          %s\n", boolToStatus(config.DNSCaching))
		fmt.Printf("    Buffer Pooling:       %s\n", boolToStatus(config.BufferPooling))
		fmt.Printf("    Keep-Alive:           %s\n", boolToStatus(config.KeepAlive))
		fmt.Printf("    Request Logging:      %s\n", boolToStatus(config.RequestLogging))
		fmt.Printf("    Profiling (pprof):    %s\n", boolToStatus(config.Profiling))
		fmt.Println()
	}

	// Startup time
	fmt.Printf("  Started:       %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Ready message
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     🚀 READY TO SERVE 🚀                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// BannerConfig holds configuration for the startup banner.
type BannerConfig struct {
	ServerAddr     string
	TLSEnabled     bool
	DatabaseType   string
	DatabasePath   string
	LogLevel       string
	ProxyCount     int
	Proxies        []ProxyInfo
	Features       map[string]bool
	ShowPerformance bool

	// Performance flags
	ConnectionPooling bool
	DNSCaching       bool
	BufferPooling    bool
	KeepAlive        bool
	RequestLogging   bool
	Profiling        bool
}

// ProxyInfo holds basic proxy information for the banner.
type ProxyInfo struct {
	Name    string
	Listen  string
	Target  string
	Enabled bool
}

// boolToStatus converts a boolean to an enabled/disabled string.
func boolToStatus(b bool) string {
	if b {
		return "Enabled  ✓"
	}
	return "Disabled ✗"
}

// PrintSimple prints a simplified banner.
func PrintSimple() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║         MODBRIDGE v" + Version + "                ║")
	fmt.Println("║   Enterprise Modbus TCP Proxy Manager    ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Println()
}

// PrintVersion prints version information only.
func PrintVersion() {
	fmt.Printf("Modbridge version %s\n", Version)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
