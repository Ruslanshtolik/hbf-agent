#!/bin/bash
set -e

# HBF Agent Installation Script

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/hbf-agent"
LOG_DIR="/var/log/hbf-agent"
SYSTEMD_DIR="/etc/systemd/system"

echo "Installing HBF Agent..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root or with sudo"
    exit 1
fi

# Check for required dependencies
echo "Checking dependencies..."
if ! command -v iptables &> /dev/null; then
    echo "Error: iptables is not installed"
    exit 1
fi

# Create directories
echo "Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"
mkdir -p "$CONFIG_DIR/certs"

# Copy binary
echo "Installing binary..."
if [ -f "./hbf-agent" ]; then
    cp ./hbf-agent "$INSTALL_DIR/hbf-agent"
    chmod +x "$INSTALL_DIR/hbf-agent"
else
    echo "Error: hbf-agent binary not found"
    exit 1
fi

# Copy configuration
echo "Installing configuration..."
if [ -f "./config/config.example.yaml" ]; then
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        cp ./config/config.example.yaml "$CONFIG_DIR/config.yaml"
        echo "Configuration file created at $CONFIG_DIR/config.yaml"
    else
        echo "Configuration file already exists, skipping..."
    fi
fi

# Copy systemd service
echo "Installing systemd service..."
if [ -f "./deploy/systemd/hbf-agent.service" ]; then
    cp ./deploy/systemd/hbf-agent.service "$SYSTEMD_DIR/hbf-agent.service"
    systemctl daemon-reload
fi

# Set permissions
echo "Setting permissions..."
chown -R root:root "$CONFIG_DIR"
chmod 755 "$CONFIG_DIR"
chmod 644 "$CONFIG_DIR/config.yaml"
chown -R root:root "$LOG_DIR"
chmod 755 "$LOG_DIR"

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit configuration: $CONFIG_DIR/config.yaml"
echo "2. Start the service: systemctl start hbf-agent"
echo "3. Enable on boot: systemctl enable hbf-agent"
echo "4. Check status: systemctl status hbf-agent"
echo "5. View logs: journalctl -u hbf-agent -f"
echo ""
