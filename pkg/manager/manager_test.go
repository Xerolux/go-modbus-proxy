package manager

import (
	"testing"

	"modbusproxy/pkg/config"
	"modbusproxy/pkg/logger"
)

func TestNewManager(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	m := NewManager(nil, log, nil)

	if m == nil {
		t.Error("NewManager returned nil")
	}

	if m.proxies == nil {
		t.Error("proxies map not initialized")
	}

	if m.broadcaster == nil {
		t.Error("broadcaster not initialized")
	}
}

func TestAddProxy(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	m := NewManager(cfgMgr, log, nil)

	cfg := config.ProxyConfig{
		ID:         "test-id",
		Name:       "Test Proxy",
		ListenAddr: ":5020",
		TargetAddr: "192.168.1.100:502",
		Enabled:    true,
	}

	err = m.AddProxy(cfg, false)
	if err != nil {
		t.Errorf("AddProxy failed: %v", err)
	}

	if len(m.proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(m.proxies))
	}

	p, ok := m.proxies["test-id"]
	if !ok {
		t.Error("Proxy not added to map")
	}

	if p.ID != "test-id" {
		t.Errorf("Expected ID test-id, got %s", p.ID)
	}
}

func TestRemoveProxy(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	m := NewManager(cfgMgr, log, nil)

	cfg := config.ProxyConfig{
		ID:         "test-id",
		Name:       "Test Proxy",
		ListenAddr: ":5020",
		TargetAddr: "192.168.1.100:502",
		Enabled:    true,
	}

	// Add proxy to config first
	err = m.AddProxy(cfg, true)
	if err != nil {
		t.Fatalf("AddProxy failed: %v", err)
	}

	err = m.RemoveProxy("test-id")
	if err != nil {
		t.Errorf("RemoveProxy failed: %v", err)
	}

	if len(m.proxies) != 0 {
		t.Errorf("Expected 0 proxies, got %d", len(m.proxies))
	}
}

func TestRemoveProxyNotFound(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	m := NewManager(cfgMgr, log, nil)

	err = m.RemoveProxy("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent proxy")
	}
}

func TestGetProxies(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	m := NewManager(cfgMgr, log, nil)

	cfg := config.ProxyConfig{
		ID:         "test-id",
		Name:       "Test Proxy",
		ListenAddr: ":5020",
		TargetAddr: "192.168.1.100:502",
		Enabled:    true,
	}

	m.AddProxy(cfg, false)

	proxies := m.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("Expected 1 proxy in list, got %d", len(proxies))
	}

	if proxies[0]["id"] != "test-id" {
		t.Errorf("Expected ID test-id, got %v", proxies[0]["id"])
	}
}

func TestGetProxyStatus(t *testing.T) {
	log, err := logger.NewLogger("test.log", 100)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()
	cfgMgr := config.NewManager("test.json")
	m := NewManager(cfgMgr, log, nil)

	cfg := config.ProxyConfig{
		ID:         "test-id",
		Name:       "Test Proxy",
		ListenAddr: ":5020",
		TargetAddr: "192.168.1.100:502",
		Enabled:    true,
		Paused:     false,
	}

	m.AddProxy(cfg, true)

	status := m.getProxyStatus("test-id")
	if status == nil {
		t.Error("getProxyStatus returned nil")
	}

	if status["id"] != "test-id" {
		t.Errorf("Expected ID test-id, got %v", status["id"])
	}

	if status["paused"] != false {
		t.Errorf("Expected paused false, got %v", status["paused"])
	}

	if status["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", status["enabled"])
	}
}
