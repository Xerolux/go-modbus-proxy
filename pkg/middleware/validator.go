package middleware

import (
	"errors"
	"modbusproxy/pkg/config"
	"net"
	"strconv"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validator handles input validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateProxyConfig validates a proxy configuration
func (v *Validator) ValidateProxyConfig(cfg config.ProxyConfig) error {
	var errs []error

	// Validate ID
	if cfg.ID == "" {
		errs = append(errs, &ValidationError{
			Field:   "id",
			Message: "cannot be empty",
		})
	}

	// Validate Name
	if cfg.Name == "" {
		errs = append(errs, &ValidationError{
			Field:   "name",
			Message: "cannot be empty",
		})
	}

	// Validate Listen Address
	if cfg.ListenAddr == "" {
		errs = append(errs, &ValidationError{
			Field:   "listen_addr",
			Message: "cannot be empty",
		})
	} else {
		if err := v.validateAddress(cfg.ListenAddr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "listen_addr",
				Message: err.Error(),
			})
		}
	}

	// Validate Target Address
	if cfg.TargetAddr == "" {
		errs = append(errs, &ValidationError{
			Field:   "target_addr",
			Message: "cannot be empty",
		})
	} else {
		if err := v.validateAddress(cfg.TargetAddr); err != nil {
			errs = append(errs, &ValidationError{
				Field:   "target_addr",
				Message: err.Error(),
			})
		}
	}

	// Validate Connection Timeout
	if cfg.ConnectionTimeout < 1 || cfg.ConnectionTimeout > 300 {
		errs = append(errs, &ValidationError{
			Field:   "connection_timeout",
			Message: "must be between 1 and 300 seconds",
		})
	}

	// Validate Read Timeout
	if cfg.ReadTimeout < 1 || cfg.ReadTimeout > 300 {
		errs = append(errs, &ValidationError{
			Field:   "read_timeout",
			Message: "must be between 1 and 300 seconds",
		})
	}

	// Validate Max Retries
	if cfg.MaxRetries < 0 || cfg.MaxRetries > 10 {
		errs = append(errs, &ValidationError{
			Field:   "max_retries",
			Message: "must be between 0 and 10",
		})
	}

	// Validate Max Read Size
	if cfg.MaxReadSize < 0 || cfg.MaxReadSize > 65535 {
		errs = append(errs, &ValidationError{
			Field:   "max_read_size",
			Message: "must be between 0 and 65535",
		})
	}

	if len(errs) > 0 {
		return v.combineErrors(errs)
	}

	return nil
}

// validateAddress validates an IP:Port address
func (v *Validator) validateAddress(addr string) error {
	// Check if it starts with : (all interfaces)
	if strings.HasPrefix(addr, ":") {
		portStr := addr[1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return errors.New("invalid port number")
		}
		if port < 1 || port > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
		return nil
	}

	// Full address (IP:Port or Host:Port)
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return errors.New("invalid address format, expected 'host:port' or ':port'")
	}

	// Validate host
	if host != "" {
		// Check if it's an IP address
		ip := net.ParseIP(host)
		if ip == nil {
			// Not an IP, check if it's a valid hostname
			if err := v.validateHostname(host); err != nil {
				return err
			}
		}
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("invalid port number")
	}
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}

// validateHostname validates a hostname
func (v *Validator) validateHostname(host string) error {
	if len(host) == 0 || len(host) > 253 {
		return errors.New("hostname must be between 1 and 253 characters")
	}

	// Check for invalid characters
	for _, c := range host {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.') {
			return errors.New("hostname contains invalid characters")
		}
	}

	return nil
}

// combineErrors combines multiple errors into one
func (v *Validator) combineErrors(errs []error) error {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return errors.New(strings.Join(messages, "; "))
}

// ValidatePort validates a port number or address string
func (v *Validator) ValidatePort(port string) error {
	if port == "" {
		return errors.New("port cannot be empty")
	}

	// Check if it starts with : (all interfaces)
	if strings.HasPrefix(port, ":") {
		portStr := port[1:]
		num, err := strconv.Atoi(portStr)
		if err != nil {
			return errors.New("invalid port number")
		}
		if num < 1 || num > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
		return nil
	}

	// Full address (IP:Port or Host:Port)
	_, portStr, err := net.SplitHostPort(port)
	if err != nil {
		return errors.New("invalid address format, expected 'host:port' or ':port'")
	}

	num, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("invalid port number")
	}
	if num < 1 || num > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}
