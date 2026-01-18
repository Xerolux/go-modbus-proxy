package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// IPAccessConfig defines access control configuration.
type IPAccessConfig struct {
	Whitelist   []string
	Blacklist   []string
	EnableCheck bool
	BanDuration time.Duration
}

// IPAccessControl provides IP-based access control.
type IPAccessControl struct {
	config IPAccessConfig
	mu     sync.RWMutex
	banned map[string]time.Time
	stats  IPAccessStats
}

// IPAccessStats tracks access control statistics.
type IPAccessStats struct {
	BlockedRequests int64
	AllowedRequests int64
	TotalRequests   int64
}

// NewIPAccessControl creates a new IP access controller.
func NewIPAccessControl(config IPAccessConfig) *IPAccessControl {
	ctl := &IPAccessControl{
		config: config,
		banned: make(map[string]time.Time),
		stats:  IPAccessStats{},
	}

	// Initialize from config
	ctl.loadFromConfig()

	return ctl
}

// loadFromConfig initializes banned IPs from blacklist.
func (ctl *IPAccessControl) loadFromConfig() {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	now := time.Now()
	for _, ip := range ctl.config.Blacklist {
		ctl.banned[ip] = now
	}
}

// Check determines if an IP is allowed access.
func (ctl *IPAccessControl) Check(ip string) bool {
	if !ctl.config.EnableCheck {
		return true
	}

	ctl.mu.RLock()
	defer ctl.mu.RUnlock()

	// Check blacklist first
	if banTime, blocked := ctl.banned[ip]; blocked {
		// Check if ban has expired
		if time.Since(banTime) > ctl.config.BanDuration {
			delete(ctl.banned, ip)
		} else {
			ctl.stats.BlockedRequests++
			return false
		}
	}

	// Check whitelist if specified
	if len(ctl.config.Whitelist) > 0 {
		for _, allowed := range ctl.config.Whitelist {
			if ip == allowed || isSubnet(ip, allowed) {
				ctl.stats.AllowedRequests++
				return true
			}
		}
	}

	return false
}

// Ban adds an IP to the blacklist temporarily.
func (ctl *IPAccessControl) Ban(ip string) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	ctl.banned[ip] = time.Now()
}

// Unban removes an IP from the blacklist.
func (ctl *IPAccessControl) Unban(ip string) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	delete(ctl.banned, ip)
}

// GetBannedList returns currently banned IPs with ban times.
func (ctl *IPAccessControl) GetBannedList() map[string]time.Time {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]time.Time, len(ctl.banned))
	for ip, t := range ctl.banned {
		result[ip] = t
	}
	return result
}

// AddToWhitelist adds an IP to the whitelist.
func (ctl *IPAccessControl) AddToWhitelist(ip string) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	// Remove from blacklist if present
	delete(ctl.banned, ip)

	ctl.config.Whitelist = append(ctl.config.Whitelist, ip)
}

// RemoveFromWhitelist removes an IP from the whitelist.
func (ctl *IPAccessControl) RemoveFromWhitelist(ip string) {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	var newWhitelist []string
	for _, allowed := range ctl.config.Whitelist {
		if allowed != ip {
			newWhitelist = append(newWhitelist, allowed)
		}
	}
	ctl.config.Whitelist = newWhitelist
}

// GetStats returns access control statistics.
func (ctl *IPAccessControl) GetStats() IPAccessStats {
	ctl.mu.RLock()
	defer ctl.mu.RUnlock()

	return IPAccessStats{
		BlockedRequests: ctl.stats.BlockedRequests,
		AllowedRequests: ctl.stats.AllowedRequests,
		TotalRequests:   ctl.stats.TotalRequests,
	}
}

// ClearExpiredBans removes expired bans.
func (ctl *IPAccessControl) ClearExpiredBans() {
	ctl.mu.Lock()
	defer ctl.mu.Unlock()

	now := time.Now()
	for ip, banTime := range ctl.banned {
		if time.Since(banTime) > ctl.config.BanDuration {
			delete(ctl.banned, ip)
		}
	}
}

// isSubnet checks if an IP is in a subnet.
func isSubnet(ip, cidr string) bool {
	// Simple implementation for /24 subnets
	if len(cidr) < 10 {
		return ip[:len(cidr)] == cidr
	}
	// For more complex CIDR, a proper library would be needed
	return false
}

// Middleware creates a middleware for IP access control.
func (ctl *IPAccessControl) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			next(w, r)
			return
		}

		// Check IP access
		if !ctl.Check(ip) {
			http.Error(w, "Forbidden: IP is not allowed", http.StatusForbidden)
			return
		}

		// Add IP to request context for later use
		ctx := context.WithValue(r.Context(), "client_ip", ip)
		next(w, r.WithContext(ctx))
	})
}
