#!/bin/bash

# Modbus Proxy Manager - Installation Script
# This script installs modbusmanager as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/modbusmanager"
BINARY_NAME="modbusmanager"
SERVICE_NAME="modbusmanager"
SERVICE_USER="modbusmanager"
DATA_DIR="/var/lib/modbusmanager"
LOG_DIR="/var/log/modbusmanager"

echo -e "${GREEN}=== Modbus Proxy Manager Installation ===${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: Please run as root (use sudo)${NC}"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}Detected architecture: $ARCH${NC}"

# Check if binary exists
if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${YELLOW}Binary not found. Building from source...${NC}"

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed. Please install Go 1.21+ first.${NC}"
        exit 1
    fi

    # Build the binary
    echo "Building $BINARY_NAME..."
    go build -ldflags="-s -w" -o $BINARY_NAME ./main.go

    if [ ! -f "$BINARY_NAME" ]; then
        echo -e "${RED}Error: Build failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}Build successful${NC}"
fi

# Stop existing service if running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "Stopping existing $SERVICE_NAME service..."
    systemctl stop $SERVICE_NAME
fi

# Create system user if doesn't exist
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "Creating system user $SERVICE_USER..."
    useradd --system --no-create-home --shell /bin/false $SERVICE_USER
fi

# Create directories
echo "Creating directories..."
mkdir -p $INSTALL_DIR
mkdir -p $DATA_DIR
mkdir -p $LOG_DIR

# Copy binary
echo "Installing binary to $INSTALL_DIR..."
cp $BINARY_NAME $INSTALL_DIR/
chmod +x $INSTALL_DIR/$BINARY_NAME

# Set ownership
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR
chown -R $SERVICE_USER:$SERVICE_USER $LOG_DIR

# Create systemd service file
echo "Creating systemd service..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=Modbus Proxy Manager
Documentation=https://github.com/Xerolux/go-modbus-proxy
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$DATA_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $LOG_DIR

# Environment
Environment="WEB_PORT=:8080"

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

# Create default config if doesn't exist
if [ ! -f "$DATA_DIR/config.json" ]; then
    echo "Creating default configuration..."
    cat > $DATA_DIR/config.json <<'EOF'
{
  "web_port": ":8080",
  "admin_pass_hash": "",
  "proxies": []
}
EOF
    chown $SERVICE_USER:$SERVICE_USER $DATA_DIR/config.json
fi

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable service
echo "Enabling $SERVICE_NAME service..."
systemctl enable $SERVICE_NAME

# Start service
echo "Starting $SERVICE_NAME service..."
systemctl start $SERVICE_NAME

# Check status
sleep 2
if systemctl is-active --quiet $SERVICE_NAME; then
    echo -e "${GREEN}=== Installation Complete ===${NC}"
    echo ""
    echo "Service Status:"
    systemctl status $SERVICE_NAME --no-pager | head -n 10
    echo ""
    echo -e "${GREEN}Modbus Proxy Manager is now running!${NC}"
    echo ""
    echo "Web Interface: http://localhost:8080"
    echo ""
    echo "Useful commands:"
    echo "  sudo systemctl status $SERVICE_NAME    - Check service status"
    echo "  sudo systemctl stop $SERVICE_NAME      - Stop service"
    echo "  sudo systemctl start $SERVICE_NAME     - Start service"
    echo "  sudo systemctl restart $SERVICE_NAME   - Restart service"
    echo "  sudo journalctl -u $SERVICE_NAME -f    - View logs"
    echo ""
    echo "Configuration: $DATA_DIR/config.json"
    echo "Logs: $LOG_DIR/"
else
    echo -e "${RED}Error: Service failed to start${NC}"
    echo "Check logs with: sudo journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi
