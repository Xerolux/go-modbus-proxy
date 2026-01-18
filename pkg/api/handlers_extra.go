package api

import (
	"encoding/json"
	"modbusproxy/pkg/config"
	"net/http"
	"runtime"
	"time"
)

func (s *Server) handleLogDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=proxy.log")
	w.Header().Set("Content-Type", "application/json")
	logs := s.log.GetRecent(10000)
	for _, l := range logs {
		json.NewEncoder(w).Encode(l)
	}
}

func (s *Server) handleConfigExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=config.json")
	w.Header().Set("Content-Type", "application/json")

	cfg := s.cfgMgr.Get()
	cfg.AdminPassHash = ""
	json.NewEncoder(w).Encode(cfg)
}

func (s *Server) handleConfigImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.cfgMgr.Update(func(c *config.Config) error {
		pass := c.AdminPassHash
		*c = newCfg
		c.AdminPassHash = pass
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mgr.Initialize()
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSystemConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfg := s.cfgMgr.Get()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
		return
	}

	if r.Method == http.MethodPut {
		var req config.Config
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := s.cfgMgr.Update(func(c *config.Config) error {
			c.LogLevel = req.LogLevel
			c.LogMaxSize = req.LogMaxSize
			c.LogMaxFiles = req.LogMaxFiles
			c.LogMaxAgeDays = req.LogMaxAgeDays
			c.TLSEnabled = req.TLSEnabled
			c.TLSCertFile = req.TLSCertFile
			c.TLSKeyFile = req.TLSKeyFile
			c.SessionTimeout = req.SessionTimeout
			c.CORSAllowedOrigins = req.CORSAllowedOrigins
			c.CORSAllowedMethods = req.CORSAllowedMethods
			c.CORSAllowedHeaders = req.CORSAllowedHeaders
			c.RateLimitEnabled = req.RateLimitEnabled
			c.RateLimitRequests = req.RateLimitRequests
			c.RateLimitBurst = req.RateLimitBurst
			c.IPWhitelistEnabled = req.IPWhitelistEnabled
			c.IPWhitelist = req.IPWhitelist
			c.IPBlacklistEnabled = req.IPBlacklistEnabled
			c.IPBlacklist = req.IPBlacklist
			c.EmailEnabled = req.EmailEnabled
			c.EmailSMTPServer = req.EmailSMTPServer
			c.EmailSMTPPort = req.EmailSMTPPort
			c.EmailFrom = req.EmailFrom
			c.EmailTo = req.EmailTo
			c.EmailUsername = req.EmailUsername
			c.EmailPassword = req.EmailPassword
			c.EmailAlertOnError = req.EmailAlertOnError
			c.EmailAlertOnWarning = req.EmailAlertOnWarning
			c.BackupEnabled = req.BackupEnabled
			c.BackupInterval = req.BackupInterval
			c.BackupRetention = req.BackupRetention
			c.BackupPath = req.BackupPath
			c.BackupDatabase = req.BackupDatabase
			c.BackupConfig = req.BackupConfig
			c.MetricsEnabled = req.MetricsEnabled
			c.MetricsPort = req.MetricsPort
			c.DebugMode = req.DebugMode
			c.MaxConnections = req.MaxConnections
			return nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

var startTime time.Time

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if startTime.IsZero() {
		startTime = time.Now()
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	proxies := s.mgr.GetProxies()
	runningProxies := 0

	for _, p := range proxies {
		if status, ok := p["status"].(string); ok && status == "Running" {
			runningProxies++
		}
	}

	info := map[string]interface{}{
		"uptime_seconds":  time.Since(startTime).Seconds(),
		"uptime_human":    time.Since(startTime).String(),
		"goroutines":      runtime.NumGoroutine(),
		"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
		"memory_sys_mb":   memStats.Sys / 1024 / 1024,
		"memory_gc_mb":    memStats.NextGC / 1024 / 1024,
		"num_cpu":         runtime.NumCPU(),
		"total_proxies":   len(proxies),
		"running_proxies": runningProxies,
		"go_version":      runtime.Version(),
		"os":              runtime.GOOS,
		"arch":            runtime.GOARCH,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}
