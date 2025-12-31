.PHONY: help build test clean run docker-build docker-run lint coverage deb-amd64 deb-arm64 deb-all

# Variables
BINARY_NAME=modbusmanager
DOCKER_IMAGE=modbus-proxy-manager
VERSION?=$(shell cat version.txt 2>/dev/null || echo "0.1.0")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"
DEB_BUILD_DIR=build
DEB_PACKAGE_NAME=modbus-proxy-manager

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./main.go

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html
	rm -rf bin/
	rm -f *.log

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker-compose up -d

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker-compose down

docker-logs: ## Show Docker container logs
	docker-compose logs -f

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

install: build ## Install the binary
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./main.go

dev: ## Run in development mode with live reload (requires air)
	@which air > /dev/null || (echo "air not installed. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

deb-amd64: ## Build .deb package for AMD64
	@echo "Building .deb package for AMD64..."
	@mkdir -p $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/{DEBIAN,opt/modbusmanager,etc/systemd/system,var/lib/modbusmanager,var/log/modbusmanager}
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/opt/modbusmanager/$(BINARY_NAME) ./main.go
	@chmod +x $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/opt/modbusmanager/$(BINARY_NAME)
	@cp modbusmanager.service $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/etc/systemd/system/
	@echo "Package: $(DEB_PACKAGE_NAME)" > $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Section: net" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Architecture: amd64" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Maintainer: Xerolux <xerolux@github.com>" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Homepage: https://github.com/Xerolux/modbridge" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Description: Modern Modbus TCP Proxy Manager with Web Interface" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Depends: systemd, adduser" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control
	@cp $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/control.tmp
	@cat INSTALL_DEBIAN.md | grep -A 100 "postinst" | grep -B 100 "exit 0" | head -n -2 | tail -n +3 > $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/postinst || true
	@chmod 755 $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64/DEBIAN/postinst || true
	@dpkg-deb --build $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64
	@mkdir -p releases
	@mv $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64.deb releases/
	@echo "✓ AMD64 package built: releases/$(DEB_PACKAGE_NAME)_$(VERSION)_amd64.deb"

deb-arm64: ## Build .deb package for ARM64
	@echo "Building .deb package for ARM64..."
	@mkdir -p $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/{DEBIAN,opt/modbusmanager,etc/systemd/system,var/lib/modbusmanager,var/log/modbusmanager}
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/opt/modbusmanager/$(BINARY_NAME) ./main.go
	@chmod +x $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/opt/modbusmanager/$(BINARY_NAME)
	@cp modbusmanager.service $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/etc/systemd/system/
	@echo "Package: $(DEB_PACKAGE_NAME)" > $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Section: net" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Priority: optional" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Architecture: arm64" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Maintainer: Xerolux <xerolux@github.com>" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Homepage: https://github.com/Xerolux/modbridge" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Description: Modern Modbus TCP Proxy Manager with Web Interface" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@echo "Depends: systemd, adduser" >> $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64/DEBIAN/control
	@dpkg-deb --build $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64
	@mkdir -p releases
	@mv $(DEB_BUILD_DIR)/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64.deb releases/
	@echo "✓ ARM64 package built: releases/$(DEB_PACKAGE_NAME)_$(VERSION)_arm64.deb"

deb-all: deb-amd64 deb-arm64 ## Build .deb packages for all architectures
	@echo "✓ All .deb packages built successfully!"
	@ls -lh releases/*.deb

.DEFAULT_GOAL := help
