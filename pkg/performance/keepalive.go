package performance

import (
	"net"
	"time"
)

// KeepAliveConfig holds TCP keep-alive configuration.
type KeepAliveConfig struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Idle     time.Duration `json:"idle" yaml:"idle"`         // Time before first keep-alive probe
	Interval time.Duration `json:"interval" yaml:"interval"` // Interval between keep-alive probes
	Count    int           `json:"count" yaml:"count"`       // Number of unacknowledged probes before connection fails
}

// DefaultKeepAliveConfig returns recommended keep-alive settings for Modbus TCP.
// These settings help detect broken connections quickly while minimizing overhead.
func DefaultKeepAliveConfig() KeepAliveConfig {
	return KeepAliveConfig{
		Enabled:  true,
		Idle:     30 * time.Second, // Start probing after 30s of idle
		Interval: 10 * time.Second, // Probe every 10s
		Count:    3,                // Give up after 3 failed probes (total 30s timeout)
	}
}

// ConfigureKeepAlive configures TCP keep-alive on a connection.
// This helps detect and clean up broken connections that would otherwise
// remain in ESTABLISHED state indefinitely.
func ConfigureKeepAlive(conn net.Conn, config KeepAliveConfig) error {
	if !config.Enabled {
		return nil
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		// Not a TCP connection, skip
		return nil
	}

	// Enable keep-alive
	if err := tcpConn.SetKeepAlive(true); err != nil {
		return err
	}

	// Set keep-alive period (time before first probe)
	if err := tcpConn.SetKeepAlivePeriod(config.Idle); err != nil {
		return err
	}

	// Note: Interval and Count require platform-specific syscalls
	// These are set via TCP_KEEPINTVL and TCP_KEEPCNT on Linux
	// For cross-platform compatibility, we only set the period here
	// Advanced users can configure these via sysctl:
	//   net.ipv4.tcp_keepalive_intvl = 10
	//   net.ipv4.tcp_keepalive_probes = 3

	return nil
}

// DialerWithKeepAlive creates a net.Dialer with keep-alive configured.
func DialerWithKeepAlive(config KeepAliveConfig) *net.Dialer {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: config.Idle, // This sets the keep-alive period
	}

	if !config.Enabled {
		dialer.KeepAlive = -1 // Disable keep-alive
	}

	return dialer
}

// OptimizedTCPConfig configures a TCP connection for optimal Modbus performance.
type OptimizedTCPConfig struct {
	NoDelay       bool          `json:"no_delay" yaml:"no_delay"`             // Disable Nagle's algorithm (recommended for Modbus)
	ReadBuffer    int           `json:"read_buffer" yaml:"read_buffer"`       // SO_RCVBUF size (0=default)
	WriteBuffer   int           `json:"write_buffer" yaml:"write_buffer"`     // SO_SNDBUF size (0=default)
	ReadTimeout   time.Duration `json:"read_timeout" yaml:"read_timeout"`     // Read timeout
	WriteTimeout  time.Duration `json:"write_timeout" yaml:"write_timeout"`   // Write timeout
	Linger        int           `json:"linger" yaml:"linger"`                 // SO_LINGER (seconds, -1=disabled)
	ReuseAddress  bool          `json:"reuse_address" yaml:"reuse_address"`   // SO_REUSEADDR
}

// DefaultTCPConfig returns optimized TCP settings for Modbus.
func DefaultTCPConfig() OptimizedTCPConfig {
	return OptimizedTCPConfig{
		NoDelay:      true,              // Disable Nagle for low-latency
		ReadBuffer:   64 * 1024,         // 64KB read buffer
		WriteBuffer:  64 * 1024,         // 64KB write buffer
		ReadTimeout:  5 * time.Second,   // 5s read timeout
		WriteTimeout: 5 * time.Second,   // 5s write timeout
		Linger:       -1,                // Disable linger (close immediately)
		ReuseAddress: true,              // Allow rapid restart
	}
}

// ConfigureTCPConnection applies optimized TCP settings to a connection.
func ConfigureTCPConnection(conn net.Conn, config OptimizedTCPConfig) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return nil // Not a TCP connection
	}

	// Disable Nagle's algorithm for low-latency
	// This is critical for Modbus which uses small request/response packets
	if config.NoDelay {
		if err := tcpConn.SetNoDelay(true); err != nil {
			return err
		}
	}

	// Set buffer sizes
	if config.ReadBuffer > 0 {
		if err := tcpConn.SetReadBuffer(config.ReadBuffer); err != nil {
			return err
		}
	}

	if config.WriteBuffer > 0 {
		if err := tcpConn.SetWriteBuffer(config.WriteBuffer); err != nil {
			return err
		}
	}

	// Set timeouts
	if config.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(config.ReadTimeout)); err != nil {
			return err
		}
	}

	if config.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout)); err != nil {
			return err
		}
	}

	// Note: SO_LINGER and SO_REUSEADDR require platform-specific syscalls
	// These are typically set on the listener socket before Accept()

	return nil
}
