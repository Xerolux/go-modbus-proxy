package api

import (
	"encoding/json"
	"fmt"
	"modbusproxy/pkg/auth"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/manager"
	"modbusproxy/pkg/metrics"
	"modbusproxy/pkg/middleware"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/google/uuid"
)

// Server is the API server.
type Server struct {
	cfgMgr      *config.Manager
	mgr         *manager.Manager
	auth        *auth.Authenticator
	log         *logger.Logger
	cors        *middleware.CORSMiddleware
	security    *middleware.SecurityMiddleware
	rateLimiter *middleware.RateLimiter
	csrf        *middleware.CSRFMiddleware
	validator   *middleware.Validator
	metrics     *metrics.Metrics
}

// NewServer creates a new API server.
func NewServer(cfg *config.Manager, mgr *manager.Manager, a *auth.Authenticator, l *logger.Logger) *Server {
	// Initialize middlewares
	corsMW := middleware.NewCORSMiddleware([]string{})
	secMW := middleware.NewSecurityMiddleware()
	rateLimiter := middleware.NewRateLimiter(60, 100) // 60 requests/minute, burst 100
	csrfMW := middleware.NewCSRFMiddleware("modbridge-csrf-secret")
	validator := middleware.NewValidator()

	return &Server{
		cfgMgr:      cfg,
		mgr:         mgr,
		auth:        a,
		log:         l,
		cors:        corsMW,
		security:    secMW,
		rateLimiter: rateLimiter,
		csrf:        csrfMW,
		validator:   validator,
		metrics:     metrics.NewMetrics(),
	}
}

// Routes registers routes.
func (s *Server) Routes(mux *http.ServeMux) {
	// Helper function to compose middlewares
	compose := func(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
		return func(handler http.HandlerFunc) http.HandlerFunc {
			for i := len(middlewares) - 1; i >= 0; i-- {
				handler = middlewares[i](handler)
			}
			return handler
		}
	}

	// Common middlewares
	publicMW := compose(s.cors.Middleware, s.security.Middleware, s.rateLimiter.Middleware)
	authMW := compose(s.cors.Middleware, s.security.Middleware, s.auth.Middleware)
	csrfMW := compose(s.cors.Middleware, s.security.Middleware, s.auth.Middleware, s.csrf.Middleware)

	// Public routes
	mux.HandleFunc("/api/health", publicMW(s.handleHealth))
	mux.HandleFunc("/api/ready", publicMW(s.handleReady))
	mux.HandleFunc("/api/status", publicMW(s.handleStatus))
	mux.HandleFunc("/api/metrics", s.cors.Middleware(s.handleMetrics))
	mux.HandleFunc("/api/login", s.cors.Middleware(s.security.Middleware(s.handleLogin)))
	mux.HandleFunc("/api/setup", s.cors.Middleware(s.security.Middleware(s.handleSetup)))

	// Pprof endpoints (debug mode only)
	if os.Getenv("DEBUG") == "true" {
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}

	// Protected routes
	mux.HandleFunc("/api/proxies", authMW(s.handleProxies))
	mux.HandleFunc("/api/proxies/stream", authMW(s.handleProxiesStream))
	mux.HandleFunc("/api/proxies/control", csrfMW(s.handleProxyControl))
	mux.HandleFunc("/api/devices", authMW(s.handleDevices))
	mux.HandleFunc("/api/devices/history", authMW(s.handleDeviceHistory))
	mux.HandleFunc("/api/logs", authMW(s.handleLogs))
	mux.HandleFunc("/api/logs/download", authMW(s.handleLogDownload))
	mux.HandleFunc("/api/logs/stream", authMW(s.handleLogStream))
	mux.HandleFunc("/api/config/export", authMW(s.handleConfigExport))
	mux.HandleFunc("/api/config/import", csrfMW(s.handleConfigImport))
	mux.HandleFunc("/api/config/webport", csrfMW(s.handleWebPort))
	mux.HandleFunc("/api/config/password", csrfMW(s.handleChangePassword))
	mux.HandleFunc("/api/config/system", csrfMW(s.handleSystemConfig))
	mux.HandleFunc("/api/system/restart", csrfMW(s.handleSystemRestart))
	mux.HandleFunc("/api/system/info", authMW(s.handleSystemInfo))
}

// handleHealth is a health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleReady is a readiness check endpoint.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ready := map[string]interface{}{
		"ready": true,
		"checks": map[string]interface{}{
			"server": "ok",
		},
	}

	// Basic readiness check - server is ready if it's running
	// More sophisticated checks can be added here (database, modbus targets, etc.)
	statusCode := http.StatusOK

	if !ready["ready"].(bool) {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ready)
}

// handleMetrics returns Prometheus metrics.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(s.metrics.GetPrometheusMetrics()))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfgMgr.Get()
	status := map[string]interface{}{
		"setup_required": cfg.AdminPassHash == "",
		"proxies":        s.mgr.GetProxies(),
	}
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := s.cfgMgr.Get()
	if cfg.AdminPassHash != "" {
		http.Error(w, "Already setup", http.StatusForbidden)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = s.cfgMgr.Update(func(c *config.Config) error {
		c.AdminPassHash = hash
		return nil
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := s.cfgMgr.Get()
	if !auth.CheckPasswordHash(req.Password, cfg.AdminPassHash) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := s.auth.CreateSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   cfg.TLSEnabled, // Only send over HTTPS if TLS is enabled
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Generate and set CSRF token
	csrfToken := s.csrf.GenerateToken(token)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   cfg.TLSEnabled,
		SameSite: http.SameSiteStrictMode,
	})

	// Return login status including force_password_change flag
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":               true,
		"force_password_change": cfg.ForcePasswordChange,
	})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := s.cfgMgr.Get()
	if !auth.CheckPasswordHash(req.CurrentPassword, cfg.AdminPassHash) {
		http.Error(w, "Invalid current password", http.StatusUnauthorized)
		return
	}

	if len(req.NewPassword) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.cfgMgr.Update(func(c *config.Config) error {
		c.AdminPassHash = hash
		c.ForcePasswordChange = false
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleProxiesStream streams proxy updates via SSE
func (s *Server) handleProxiesStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := s.mgr.GetProxyEventsSubscription()
	defer s.mgr.UnsubscribeProxyEvents(ch)

	timeout := time.NewTimer(30 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout.C:
			s.log.Info("SSE stream timeout, closing connection", "")
			return
		case event := <-ch:
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			timeout.Reset(30 * time.Minute)
		}
	}
}

func (s *Server) handleProxies(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_ = json.NewEncoder(w).Encode(s.mgr.GetProxies())
		return
	}

	if r.Method == http.MethodPost {
		var req config.ProxyConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ID == "" {
			req.ID = uuid.New().String()
		}

		if err := s.validator.ValidateProxyConfig(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.mgr.AddProxy(req, true); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.Enabled && !req.Paused {
			_ = s.mgr.StartProxy(req.ID)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodPut {
		var req config.ProxyConfig
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.validator.ValidateProxyConfig(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.mgr.UpdateProxy(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodDelete {
		id := r.URL.Query().Get("id")
		if err := s.mgr.RemoveProxy(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *Server) handleProxyControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID     string `json:"id"`
		Action string `json:"action"` // start, stop, restart, pause, resume, start_all, stop_all
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	switch req.Action {
	case "start":
		err = s.mgr.StartProxy(req.ID)
	case "stop":
		err = s.mgr.StopProxy(req.ID)
	case "restart":
		_ = s.mgr.StopProxy(req.ID)
		time.Sleep(100 * time.Millisecond)
		err = s.mgr.StartProxy(req.ID)
	case "pause":
		err = s.mgr.PauseProxy(req.ID)
	case "resume":
		err = s.mgr.ResumeProxy(req.ID)
	case "start_all":
		s.mgr.StartAll()
	case "stop_all":
		s.mgr.StopAll()
	default:
		err = fmt.Errorf("unknown action")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	logs := s.log.GetRecent(100)
	_ = json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.log.Subscribe()
	defer s.log.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case entry := <-ch:
			data, _ := json.Marshal(entry)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *Server) handleSystemRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.log.Info("System restart requested via API", "")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"restarting"}`))

	// Stop all proxies gracefully
	go func() {
		time.Sleep(500 * time.Millisecond)
		s.mgr.StopAll()
		time.Sleep(500 * time.Millisecond)
		s.log.Info("System restarting now", "")
		// Exit with code 0 so the process manager (systemd, docker, etc.) can restart it
		// Note: This assumes the service is running under a process manager
		panic("Restart requested") // This will trigger main's recover and graceful shutdown
	}()
}

func (s *Server) handleWebPort(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfg := s.cfgMgr.Get()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"web_port": cfg.WebPort})
		return
	}

	if r.Method == http.MethodPut {
		var req struct {
			WebPort string `json:"web_port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate port format
		if req.WebPort == "" {
			http.Error(w, "web_port cannot be empty", http.StatusBadRequest)
			return
		}

		if err := s.validator.ValidatePort(req.WebPort); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update config
		if err := s.cfgMgr.Update(func(c *config.Config) error {
			c.WebPort = req.WebPort
			return nil
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Port updated, restart required"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
