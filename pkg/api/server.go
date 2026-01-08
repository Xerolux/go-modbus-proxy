package api

import (
	"encoding/json"
	"fmt"
	"modbusproxy/pkg/auth"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/manager"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Server is the API server.
type Server struct {
	cfgMgr *config.Manager
	mgr    *manager.Manager
	auth   *auth.Authenticator
	log    *logger.Logger
}

// NewServer creates a new API server.
func NewServer(cfg *config.Manager, mgr *manager.Manager, a *auth.Authenticator, l *logger.Logger) *Server {
	return &Server{
		cfgMgr: cfg,
		mgr:    mgr,
		auth:   a,
		log:    l,
	}
}

// Routes registers routes.
func (s *Server) Routes(mux *http.ServeMux) {
	// Public
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.corsMiddleware(s.handleStatus))
	mux.HandleFunc("/api/login", s.corsMiddleware(s.handleLogin))
	mux.HandleFunc("/api/setup", s.corsMiddleware(s.handleSetup))

	// Protected
	mux.HandleFunc("/api/proxies", s.corsMiddleware(s.auth.Middleware(s.handleProxies)))
	mux.HandleFunc("/api/proxies/control", s.corsMiddleware(s.auth.Middleware(s.handleProxyControl)))
	mux.HandleFunc("/api/devices", s.corsMiddleware(s.auth.Middleware(s.handleDevices)))
	mux.HandleFunc("/api/devices/history", s.corsMiddleware(s.auth.Middleware(s.handleDeviceHistory)))
	mux.HandleFunc("/api/logs", s.corsMiddleware(s.auth.Middleware(s.handleLogs)))
	mux.HandleFunc("/api/logs/download", s.corsMiddleware(s.auth.Middleware(s.handleLogDownload)))
	mux.HandleFunc("/api/logs/stream", s.corsMiddleware(s.auth.Middleware(s.handleLogStream)))
	mux.HandleFunc("/api/config/export", s.corsMiddleware(s.auth.Middleware(s.handleConfigExport)))
	mux.HandleFunc("/api/config/import", s.corsMiddleware(s.auth.Middleware(s.handleConfigImport)))
	mux.HandleFunc("/api/config/webport", s.corsMiddleware(s.auth.Middleware(s.handleWebPort)))
	mux.HandleFunc("/api/system/restart", s.corsMiddleware(s.auth.Middleware(s.handleSystemRestart)))
}

// corsMiddleware adds CORS headers to responses.
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleHealth is a health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
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
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
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

		// Set defaults for new fields if not provided
		if req.ConnectionTimeout == 0 {
			req.ConnectionTimeout = 10
		}
		if req.ReadTimeout == 0 {
			req.ReadTimeout = 30
		}
		if req.MaxRetries == 0 {
			req.MaxRetries = 3
		}

		if err := s.mgr.AddProxy(req, true); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If enabled and not paused, start it
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

		// Set defaults for new fields if not provided
		if req.ConnectionTimeout == 0 {
			req.ConnectionTimeout = 10
		}
		if req.ReadTimeout == 0 {
			req.ReadTimeout = 30
		}
		if req.MaxRetries == 0 {
			req.MaxRetries = 3
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
		Action string `json:"action"` // start, stop, restart, pause, resume
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
