package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"modbusproxy/pkg/config"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/manager"
	"modbusproxy/pkg/middleware"
)

func TestHandleHealth(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	server := NewServer(config.NewManager("test.json"), manager.NewManager(config.NewManager("test.json"), log, nil), nil, log)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", response["status"])
	}
}

func TestHandleStatus(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	mgr := manager.NewManager(cfgMgr, log, nil)
	server := NewServer(cfgMgr, mgr, nil, log)

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["setup_required"] != true {
		t.Error("Expected setup_required true")
	}

	proxies, ok := response["proxies"].([]interface{})
	if !ok {
		t.Fatalf("proxies is not a list, got %T: %v", response["proxies"], response["proxies"])
	}

	if len(proxies) != 0 {
		t.Errorf("Expected 0 proxies, got %d", len(proxies))
	}
}

func TestHandleProxiesGet(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	mgr := manager.NewManager(cfgMgr, log, nil)
	server := NewServer(cfgMgr, mgr, nil, log)

	req := httptest.NewRequest("GET", "/api/proxies", nil)
	w := httptest.NewRecorder()

	server.handleProxies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var proxies []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &proxies)

	if len(proxies) != 0 {
		t.Errorf("Expected 0 proxies, got %d", len(proxies))
	}
}

func TestHandleProxiesPostInvalid(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	mgr := manager.NewManager(cfgMgr, log, nil)
	server := NewServer(cfgMgr, mgr, nil, log)

	req := httptest.NewRequest("POST", "/api/proxies", nil)
	w := httptest.NewRecorder()

	server.handleProxies(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleProxiesPostValid(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	mgr := manager.NewManager(cfgMgr, log, nil)
	server := NewServer(cfgMgr, mgr, nil, log)

	cfg := config.ProxyConfig{
		ID:                "test-id",
		Name:              "Test Proxy",
		ListenAddr:        ":5020",
		TargetAddr:        "192.168.1.100:502",
		ConnectionTimeout: 5,
		ReadTimeout:       5,
		MaxRetries:        3,
		MaxReadSize:       100,
	}

	body, _ := json.Marshal(cfg)
	req := httptest.NewRequest("POST", "/api/proxies", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.handleProxies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	proxies := mgr.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(proxies))
	}
}

func TestMiddlewareChain(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	secMW := middleware.NewSecurityMiddleware()
	corsMW := middleware.NewCORSMiddleware([]string{})

	called := false
	handler := secMW.Middleware(corsMW.Middleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !called {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header missing")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options header missing")
	}
}
