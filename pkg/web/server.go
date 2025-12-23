package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"modbusproxy/pkg/config"
	"net/http"
)

//go:embed templates/index.html
var templateFS embed.FS

// Server is the web server.
type Server struct {
	cfgManager *config.Manager
	restartCb  func(config.Config)
}

// NewServer creates a new web server.
func NewServer(cfgManager *config.Manager, restartCb func(config.Config)) *Server {
	return &Server{
		cfgManager: cfgManager,
		restartCb:  restartCb,
	}
}

// Start starts the web server.
func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/status", s.handleStatus)

	addr := s.cfgManager.Get().WebPort
	log.Printf("Web interface listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Web server error: %v", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfg := s.cfgManager.Get()
		json.NewEncoder(w).Encode(cfg)
	} else if r.Method == http.MethodPost {
		var newConfig config.Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.cfgManager.Update(func(c *config.Config) error {
			*c = newConfig
			return nil
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Trigger restart in background
		go s.restartCb(newConfig)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status": "Running",
	}
	json.NewEncoder(w).Encode(status)
}
