#!/bin/sh
set -e

INSTALL_BIN="/usr/local/bin/dnf"
INSTALL_SERVICE="/etc/systemd/system/dnf.service"
INSTALL_CONFIG="/etc/dnf/config.yml"

REPO="robinhickmann/dnf"
VERSION=${1:-latest}

# Check for root and systemd
if [ "$(id -u)" -ne 0 ]; then
    echo "error: this script must be run as root"
    exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
    echo "error: this script requires systemd"
    exit 1
fi

# Detect OS/ARCH and get the latest version
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *) echo "error: unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' \
        | cut -d'"' -f4)
fi

RELEASE_URL="https://github.com/$REPO/releases/download/$VERSION/"
echo "Installing dnf $VERSION ($OS/$ARCH)"

# Download and install binary
if [ ! -f "$INSTALL_BIN" ]; then
    curl -fsSL "$RELEASE_URL/dnf-$OS-$ARCH" \
        -o "$INSTALL_BIN"
fi

# Download and install systemd service
if [ ! -f "$INSTALL_SERVICE" ]; then
    curl -fsSL "$RELEASE_URL/dnf.service" \
        -o "$INSTALL_SERVICE"
fi

# install config only if it doesn't exist
if [ ! -f "$INSTALL_CONFIG" ]; then
    mkdir -p /etc/dnf

    curl -fsSL "$RELEASE_URL/config.yml" \
        -o "$INSTALL_CONFIG"
fi

# Enable and start service
systemctl daemon-reload
systemctl enable --now dnf

echo "Successfully installed dnf $VERSION ($OS/$ARCH)"
