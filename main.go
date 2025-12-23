package main

import (
	"log"
	"modbusproxy/pkg/config"
	"modbusproxy/pkg/proxy"
	"modbusproxy/pkg/web"
	"sync"
)

func main() {
	// Initialize config
	cfgManager := config.NewManager("config.json")
	if err := cfgManager.Load(); err != nil {
		log.Printf("Could not load config, using defaults: %v", err)
	}

	// Manage proxy instance
	var proxyMu sync.Mutex
	var currentProxy *proxy.Proxy

	startProxy := func(cfg config.Config) {
		proxyMu.Lock()
		defer proxyMu.Unlock()

		if currentProxy != nil {
			log.Println("Stopping existing proxy...")
			currentProxy.Stop()
		}

		log.Println("Starting new proxy...")
		currentProxy = proxy.NewProxy(cfg)
		go func(p *proxy.Proxy) {
			if err := p.Start(); err != nil {
				log.Printf("Proxy server error: %v", err)
			}
		}(currentProxy)
	}

	// Initial start
	startProxy(cfgManager.Get())

	// Start web server
	webServer := web.NewServer(cfgManager, func(newConfig config.Config) {
		startProxy(newConfig)
	})
	
	// This blocks
	webServer.Start()
}
