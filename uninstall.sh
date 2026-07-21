#!/bin/sh
set -e

INSTALL_BIN="/usr/local/bin/dnfd"
INSTALL_SERVICE="/etc/systemd/system/dnfd.service"
INSTALL_DIR="/etc/dnfd"

# Check for root and systemd
if [ "$(id -u)" -ne 0 ]; then
    echo "error: this script must be run as root"
    exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
    echo "error: this script requires systemd"
    exit 1
fi

if ! { [ -f "$INSTALL_BIN" ] || [ -f "$INSTALL_SERVICE" ] || [ -d "$INSTALL_DIR" ]; }; then
    echo "dnf is not installed"
    exit 0
fi

if [ -f "$INSTALL_SERVICE" ]; then
    systemctl disable --now dnfd
    rm -f "$INSTALL_SERVICE"
    systemctl daemon-reload
fi

# Remove files
rm -rf $INSTALL_BIN $INSTALL_DIR

echo "Successfully uninstalled dnf"
