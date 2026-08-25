#!/bin/sh
set -e

INSTALL_BIN="/usr/local/bin/dnfd"
INSTALL_SERVICE="/etc/systemd/system/dnfd.service"
INSTALL_DIR="/etc/dnfd"
INSTALL_USER="dnfd"

LOCAL_VERSION=$([ -x $INSTALL_BIN ] && $INSTALL_BIN --version \
    | awk '/Version:/ {print $2}' || echo 0)

ask() {
    question=$1
    while true; do
        printf "%s [y/N] " "$question"
        read -r answer < /dev/tty
        case $answer in
            [Yy]*) return 0 ;;
            [Nn]*|"") return 1 ;;
        esac
    done
}

verify_system() {
    if ! command -v systemctl >/dev/null 2>&1; then
        echo "Error: This script requires systemd"
        exit 1
    fi
}

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
fi

verify_system

if ! { [ -f "$INSTALL_BIN" ] || [ -f "$INSTALL_SERVICE" ] || [ -d "$INSTALL_DIR" ]; }; then
    echo "Error: dnf is not installed"
    exit 1
fi

if ask "Are you sure you want to uninstall dnf $LOCAL_VERSION ($OS/$ARCH)?"; then
    echo "Uninstalling dnf $LOCAL_VERSION ($OS/$ARCH)"
else
    echo "Installation cancelled."
    exit 0
fi

if [ -f "$INSTALL_SERVICE" ]; then
    $SUDO systemctl disable --now dnfd
    $SUDO rm -f "$INSTALL_SERVICE"
    $SUDO systemctl daemon-reload
fi

$SUDO rm -rf $INSTALL_BIN $INSTALL_DIR
$SUDO userdel $INSTALL_USER

echo "Successfully uninstalled dnf"
