package main

import (
	"context"
	"log"
	"modbusproxy/pkg/api"
	"modbusproxy/pkg/auth"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/database"
	"modbusproxy/pkg/logger"
	"modbusproxy/pkg/manager"
	"modbusproxy/pkg/web"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Database
	db, err := database.NewDB("modbridge.db")
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	// 2. Config
	cfgMgr := config.NewManager("config.json")
	if err := cfgMgr.Load(); err != nil {
		log.Printf("Starting with empty config: %v", err)
	}

	// 3. Logger
	l, err := logger.NewLogger("proxy.log", 1000)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer l.Close()

	// 4. Manager
	mgr := manager.NewManager(cfgMgr, l, db)
	mgr.Initialize()

	// 4. Auth
	authenticator := auth.NewAuthenticator()
	go authenticator.CleanupExpiredSessions()

	// 5. API Server
	apiServer := api.NewServer(cfgMgr, mgr, authenticator, l)

	// 6. Router
	mux := http.NewServeMux()

	// API Routes
	apiServer.Routes(mux)

	// Web Routes
	mux.Handle("/", web.Handler())

	// Start server
	addr := cfgMgr.Get().WebPort
	if addr == "" {
		addr = ":8080"
	}
	// Override with environment variable if set
	if envPort := os.Getenv("WEB_PORT"); envPort != "" {
		addr = envPort
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	l.Info("SYSTEM", "Starting Modbus Manager on "+addr)
	log.Printf("Listening on %s", addr)

	// Run server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("SYSTEM", "Shutting down server...")
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop all proxies
	mgr.StopAll()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	l.Info("SYSTEM", "Server stopped")
	log.Println("Server stopped")
}
