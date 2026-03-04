#!/bin/sh
# Pre-remove script for gflow package

set -e

# Stop services before removal
if command -v systemctl > /dev/null 2>&1; then
    systemctl stop gflow-server 2>/dev/null || true
    systemctl stop gflow-runner 2>/dev/null || true
fi

echo "GFlow services stopped."
