#!/bin/sh
# Post-install script for gflow package

set -e

# Create gflow user and group if they don't exist
if ! getent passwd gflow > /dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false gflow
fi

# Create directories
mkdir -p /var/lib/gflow
mkdir -p /var/log/gflow
mkdir -p /etc/gflow

# Set permissions
chown -R gflow:gflow /var/lib/gflow
chown -R gflow:gflow /var/log/gflow

echo "GFlow has been installed successfully."
echo "Configuration files are in /etc/gflow/"
echo "To start the server: systemctl start gflow-server"
echo "To start the runner: systemctl start gflow-runner"
