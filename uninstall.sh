#!/bin/sh
set -e

INSTALL_BIN="/usr/local/bin/dnf"
INSTALL_SERVICE="/etc/systemd/system/dnf.service"
INSTALL_DIR="/etc/dnf"

# Check for root and systemd
if [ "$(id -u)" -ne 0 ]; then
    echo "error: this script must be run as root"
    exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
    echo "error: this script requires systemd"
    exit 1
fi

# Disable and stop service
systemctl disable --now dnf

# Remove files
rm -rf $INSTALL_BIN $INSTALL_SERVICE $INSTALL_DIR

systemctl daemon-reload

echo "Successfully uninstalled dnf"
