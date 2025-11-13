#!/usr/bin/env bash

# EvilGoPhish Stop Script
# This script stops all components gracefully

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_DIR/shutdown_${TIMESTAMP}.log"
}

log "Stopping EvilGoPhish components..."

# Function to stop a process by PID file
stop_process() {
    local PID_FILE=$1
    local NAME=$2
    
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null 2>&1; then
            log "Stopping $NAME (PID: $PID)..."
            kill -TERM $PID
            
            # Wait for graceful shutdown (max 10 seconds)
            for i in {1..10}; do
                if ! ps -p $PID > /dev/null 2>&1; then
                    log "$NAME stopped successfully"
                    rm -f "$PID_FILE"
                    return 0
                fi
                sleep 1
            done
            
            # Force kill if still running
            if ps -p $PID > /dev/null 2>&1; then
                log "Force killing $NAME..."
                kill -9 $PID
                sleep 1
            fi
        else
            log "$NAME is not running (PID $PID not found)"
        fi
        rm -f "$PID_FILE"
    else
        log "No PID file found for $NAME"
    fi
}

# Stop components in reverse order
stop_process "$SCRIPT_DIR/evilginx3.pid" "Evilginx3"
stop_process "$SCRIPT_DIR/evilfeed.pid" "EvilFeed"
stop_process "$SCRIPT_DIR/gophish.pid" "GoPhish"

# Also try to kill by process name as backup
pkill -f "evilginx3" 2>/dev/null || true
pkill -f "evilfeed" 2>/dev/null || true
pkill -f "gophish" 2>/dev/null || true

log "All components stopped"
