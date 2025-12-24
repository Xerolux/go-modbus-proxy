package logging

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// RequestLogger logs Modbus requests and responses for debugging and auditing.
type RequestLogger struct {
	mu      sync.RWMutex
	enabled bool
	verbose bool
	logFile string
	logger  *log.Logger

	// Stats
	logged uint64
}

// Config holds request logger configuration.
type Config struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Verbose bool   `json:"verbose" yaml:"verbose"` // Log full frame data
	LogFile string `json:"log_file" yaml:"log_file"` // Optional log file
}

// NewRequestLogger creates a new request logger.
func NewRequestLogger(config Config) *RequestLogger {
	rl := &RequestLogger{
		enabled: config.Enabled,
		verbose: config.Verbose,
		logFile: config.LogFile,
		logger:  log.Default(),
	}

	// TODO: Open log file if specified
	// For now, just use default logger

	return rl
}

// LogRequest logs a Modbus request.
func (rl *RequestLogger) LogRequest(conn net.Conn, data []byte, proxyID string) {
	if !rl.enabled {
		return
	}

	rl.mu.Lock()
	rl.logged++
	rl.mu.Unlock()

	if rl.verbose {
		rl.logger.Printf("[REQUEST] [%s] from=%s len=%d data=%x",
			proxyID, conn.RemoteAddr(), len(data), data)
	} else {
		// Parse basic Modbus info
		if len(data) >= 8 {
			transID := uint16(data[0])<<8 | uint16(data[1])
			unitID := data[6]
			funcCode := data[7]
			rl.logger.Printf("[REQUEST] [%s] from=%s trans=%d unit=%d func=0x%02x len=%d",
				proxyID, conn.RemoteAddr(), transID, unitID, funcCode, len(data))
		} else {
			rl.logger.Printf("[REQUEST] [%s] from=%s len=%d (invalid)",
				proxyID, conn.RemoteAddr(), len(data))
		}
	}
}

// LogResponse logs a Modbus response.
func (rl *RequestLogger) LogResponse(conn net.Conn, data []byte, proxyID string, latency time.Duration) {
	if !rl.enabled {
		return
	}

	if rl.verbose {
		rl.logger.Printf("[RESPONSE] [%s] to=%s len=%d latency=%s data=%x",
			proxyID, conn.RemoteAddr(), len(data), latency, data)
	} else {
		// Parse basic Modbus info
		if len(data) >= 8 {
			transID := uint16(data[0])<<8 | uint16(data[1])
			unitID := data[6]
			funcCode := data[7]

			// Check for exception
			isException := (funcCode & 0x80) != 0
			if isException {
				exceptionCode := uint8(0)
				if len(data) >= 9 {
					exceptionCode = data[8]
				}
				rl.logger.Printf("[RESPONSE] [%s] to=%s trans=%d unit=%d EXCEPTION func=0x%02x code=0x%02x latency=%s",
					proxyID, conn.RemoteAddr(), transID, unitID, funcCode&0x7F, exceptionCode, latency)
			} else {
				rl.logger.Printf("[RESPONSE] [%s] to=%s trans=%d unit=%d func=0x%02x len=%d latency=%s",
					proxyID, conn.RemoteAddr(), transID, unitID, funcCode, len(data), latency)
			}
		} else {
			rl.logger.Printf("[RESPONSE] [%s] to=%s len=%d latency=%s (invalid)",
				proxyID, conn.RemoteAddr(), len(data), latency)
		}
	}
}

// LogError logs an error.
func (rl *RequestLogger) LogError(conn net.Conn, err error, proxyID string) {
	if !rl.enabled {
		return
	}

	from := "unknown"
	if conn != nil {
		from = conn.RemoteAddr().String()
	}

	rl.logger.Printf("[ERROR] [%s] from=%s error=%v", proxyID, from, err)
}

// LogConnection logs a new connection.
func (rl *RequestLogger) LogConnection(conn net.Conn, proxyID string) {
	if !rl.enabled {
		return
	}

	rl.logger.Printf("[CONNECT] [%s] from=%s", proxyID, conn.RemoteAddr())
}

// LogDisconnection logs a disconnection.
func (rl *RequestLogger) LogDisconnection(conn net.Conn, proxyID string, duration time.Duration) {
	if !rl.enabled {
		return
	}

	rl.logger.Printf("[DISCONNECT] [%s] from=%s duration=%s",
		proxyID, conn.RemoteAddr(), duration)
}

// Enable enables request logging.
func (rl *RequestLogger) Enable() {
	rl.mu.Lock()
	rl.enabled = true
	rl.mu.Unlock()
}

// Disable disables request logging.
func (rl *RequestLogger) Disable() {
	rl.mu.Lock()
	rl.enabled = false
	rl.mu.Unlock()
}

// SetVerbose sets verbose mode.
func (rl *RequestLogger) SetVerbose(verbose bool) {
	rl.mu.Lock()
	rl.verbose = verbose
	rl.mu.Unlock()
}

// IsEnabled returns whether logging is enabled.
func (rl *RequestLogger) IsEnabled() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.enabled
}

// Stats returns logging statistics.
func (rl *RequestLogger) Stats() uint64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.logged
}

// ParseModbusFrame parses basic information from a Modbus TCP frame.
func ParseModbusFrame(data []byte) (transID uint16, unitID uint8, funcCode uint8, err error) {
	if len(data) < 8 {
		err = fmt.Errorf("frame too short: %d bytes", len(data))
		return
	}

	transID = uint16(data[0])<<8 | uint16(data[1])
	// data[2:4] is protocol ID (should be 0)
	// data[4:6] is length
	unitID = data[6]
	funcCode = data[7]

	return
}

// GlobalRequestLogger is the global request logger instance.
var GlobalRequestLogger = NewRequestLogger(Config{
	Enabled: false, // Disabled by default for performance
	Verbose: false,
})
