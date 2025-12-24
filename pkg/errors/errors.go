package errors

import (
	"errors"
	"fmt"
)

// Error types for better error handling and user-friendly messages.
var (
	// Configuration errors
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrMissingConfig     = errors.New("configuration file not found")
	ErrInvalidListenAddr = errors.New("invalid listen address")
	ErrInvalidTargetAddr = errors.New("invalid target address")
	ErrPortInUse         = errors.New("port already in use")
	ErrPortConflict      = errors.New("port conflict detected")

	// Connection errors
	ErrConnectionRefused  = errors.New("connection refused")
	ErrConnectionTimeout  = errors.New("connection timeout")
	ErrConnectionReset    = errors.New("connection reset by peer")
	ErrConnectionClosed   = errors.New("connection closed")
	ErrTooManyConnections = errors.New("too many connections")
	ErrNoAvailableConn    = errors.New("no available connections in pool")

	// Modbus protocol errors
	ErrInvalidFrame       = errors.New("invalid Modbus frame")
	ErrFrameTooShort      = errors.New("Modbus frame too short")
	ErrInvalidTransID     = errors.New("invalid transaction ID")
	ErrInvalidProtocolID  = errors.New("invalid protocol ID")
	ErrInvalidLength      = errors.New("invalid length field")
	ErrInvalidUnitID      = errors.New("invalid unit ID")
	ErrInvalidFunctionCode = errors.New("invalid function code")
	ErrModbusException    = errors.New("Modbus exception response")

	// Proxy errors
	ErrProxyNotFound   = errors.New("proxy not found")
	ErrProxyNotRunning = errors.New("proxy not running")
	ErrProxyExists     = errors.New("proxy already exists")
	ErrProxyDisabled   = errors.New("proxy is disabled")
	ErrProxyStartFailed = errors.New("failed to start proxy")
	ErrProxyStopFailed  = errors.New("failed to stop proxy")

	// Authentication/authorization errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrInsufficientPerms = errors.New("insufficient permissions")

	// Resource errors
	ErrResourceNotFound    = errors.New("resource not found")
	ErrResourceExists      = errors.New("resource already exists")
	ErrResourceUnavailable = errors.New("resource temporarily unavailable")
	ErrQuotaExceeded       = errors.New("quota exceeded")

	// System errors
	ErrDatabaseError   = errors.New("database error")
	ErrInternalError   = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ModbusError represents a Modbus exception response.
type ModbusError struct {
	FunctionCode  uint8
	ExceptionCode uint8
	Message       string
}

func (e *ModbusError) Error() string {
	return fmt.Sprintf("Modbus exception: function 0x%02X, exception 0x%02X - %s",
		e.FunctionCode, e.ExceptionCode, e.Message)
}

// NewModbusError creates a new Modbus error.
func NewModbusError(functionCode, exceptionCode uint8) *ModbusError {
	return &ModbusError{
		FunctionCode:  functionCode,
		ExceptionCode: exceptionCode,
		Message:       getExceptionMessage(exceptionCode),
	}
}

// getExceptionMessage returns a human-readable message for Modbus exception codes.
func getExceptionMessage(code uint8) string {
	switch code {
	case 0x01:
		return "Illegal Function"
	case 0x02:
		return "Illegal Data Address"
	case 0x03:
		return "Illegal Data Value"
	case 0x04:
		return "Slave Device Failure"
	case 0x05:
		return "Acknowledge"
	case 0x06:
		return "Slave Device Busy"
	case 0x08:
		return "Memory Parity Error"
	case 0x0A:
		return "Gateway Path Unavailable"
	case 0x0B:
		return "Gateway Target Device Failed to Respond"
	default:
		return fmt.Sprintf("Unknown Exception (0x%02X)", code)
	}
}

// ConfigError represents a configuration error with context.
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in '%s': %s (value: %s)", e.Field, e.Message, e.Value)
}

// NewConfigError creates a new configuration error.
func NewConfigError(field, value, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ProxyError represents a proxy-specific error with context.
type ProxyError struct {
	ProxyID string
	ProxyName string
	Err     error
}

func (e *ProxyError) Error() string {
	if e.ProxyName != "" {
		return fmt.Sprintf("proxy '%s' (%s): %v", e.ProxyName, e.ProxyID, e.Err)
	}
	return fmt.Sprintf("proxy '%s': %v", e.ProxyID, e.Err)
}

func (e *ProxyError) Unwrap() error {
	return e.Err
}

// NewProxyError creates a new proxy error.
func NewProxyError(proxyID, proxyName string, err error) *ProxyError {
	return &ProxyError{
		ProxyID:   proxyID,
		ProxyName: proxyName,
		Err:       err,
	}
}

// ConnectionError represents a connection error with details.
type ConnectionError struct {
	RemoteAddr string
	LocalAddr  string
	Err        error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error (remote=%s, local=%s): %v",
		e.RemoteAddr, e.LocalAddr, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewConnectionError creates a new connection error.
func NewConnectionError(remoteAddr, localAddr string, err error) *ConnectionError {
	return &ConnectionError{
		RemoteAddr: remoteAddr,
		LocalAddr:  localAddr,
		Err:        err,
	}
}

// Wrap wraps an error with additional context.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapConfig wraps an error with configuration context.
func WrapConfig(err error, field, value string) error {
	if err == nil {
		return nil
	}
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: err.Error(),
	}
}

// WrapProxy wraps an error with proxy context.
func WrapProxy(err error, proxyID, proxyName string) error {
	if err == nil {
		return nil
	}
	return NewProxyError(proxyID, proxyName, err)
}

// WrapConnection wraps an error with connection context.
func WrapConnection(err error, remoteAddr, localAddr string) error {
	if err == nil {
		return nil
	}
	return NewConnectionError(remoteAddr, localAddr, err)
}

// IsModbusException checks if an error is a Modbus exception.
func IsModbusException(err error) bool {
	var modbusErr *ModbusError
	return errors.As(err, &modbusErr)
}

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	var cfgErr *ConfigError
	return errors.As(err, &cfgErr)
}

// IsProxyError checks if an error is a proxy error.
func IsProxyError(err error) bool {
	var proxyErr *ProxyError
	return errors.As(err, &proxyErr)
}

// IsConnectionError checks if an error is a connection error.
func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr)
}

// HTTPStatus returns an appropriate HTTP status code for an error.
func HTTPStatus(err error) int {
	if err == nil {
		return 200
	}

	switch {
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidCredentials):
		return 401
	case errors.Is(err, ErrInsufficientPerms):
		return 403
	case errors.Is(err, ErrResourceNotFound), errors.Is(err, ErrProxyNotFound):
		return 404
	case errors.Is(err, ErrResourceExists), errors.Is(err, ErrProxyExists), errors.Is(err, ErrPortInUse):
		return 409
	case errors.Is(err, ErrTooManyConnections), errors.Is(err, ErrQuotaExceeded):
		return 429
	case errors.Is(err, ErrServiceUnavailable), errors.Is(err, ErrResourceUnavailable):
		return 503
	case IsConfigError(err):
		return 400
	case errors.Is(err, ErrInternalError):
		return 500
	default:
		return 500
	}
}

// UserMessage returns a user-friendly error message.
func UserMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for specific error types
	var modbusErr *ModbusError
	if errors.As(err, &modbusErr) {
		return modbusErr.Message
	}

	var cfgErr *ConfigError
	if errors.As(err, &cfgErr) {
		return fmt.Sprintf("Configuration error in '%s': %s", cfgErr.Field, cfgErr.Message)
	}

	var proxyErr *ProxyError
	if errors.As(err, &proxyErr) {
		return fmt.Sprintf("Proxy '%s': %v", proxyErr.ProxyName, proxyErr.Err)
	}

	// Check for known errors
	switch {
	case errors.Is(err, ErrConnectionRefused):
		return "Unable to connect to target server. Please check if the target is running and accessible."
	case errors.Is(err, ErrConnectionTimeout):
		return "Connection timed out. The target server may be overloaded or unreachable."
	case errors.Is(err, ErrPortInUse):
		return "Port is already in use. Please choose a different port or stop the conflicting service."
	case errors.Is(err, ErrTooManyConnections):
		return "Too many connections. Please try again later or increase connection limits."
	case errors.Is(err, ErrUnauthorized):
		return "Authentication required. Please provide valid credentials."
	case errors.Is(err, ErrInsufficientPerms):
		return "You don't have permission to perform this action."
	default:
		return err.Error()
	}
}
