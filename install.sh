#!/bin/sh
set -eo pipefail

REPO="robinhickmann/dnf"

INSTALL_BIN="/usr/local/bin/dnfd"
INSTALL_SERVICE="/etc/systemd/system/dnfd.service"
INSTALL_CONFIG="/etc/dnfd/config.yml"
INSTALL_USER="dnfd"

VERSION=${1:-latest}
LOCAL_VERSION=$([ -x $INSTALL_BIN ] && $INSTALL_BIN --version \
    | awk '/Version:/ {print $2}' || echo 0)

verify_system() {
    if ! command -v systemctl >/dev/null 2>&1; then
        echo "Error: This script requires systemd"
        exit 1
    fi
}

verify_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64) ARCH="arm64" ;;
        *) echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
    esac
}

ask() {
    question=$1

    if [ "$YES" = "1" ]; then
        return 0
    fi

    while true; do
        printf "%s [y/N] " "$question"
        read -r answer < /dev/tty
        case $answer in
            [Yy]*) return 0 ;;
            [Nn]*|"") return 1 ;;
        esac
    done
}

compare_version() {
    if [ "$1" = "$2" ]; then
        echo "same"
        return
    fi

    if [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -n1)" = "$2" ]; then
        echo "higher"
    else
        echo "lower"
    fi
}

setup_tmp() {
    TMP_DIR=$(mktemp -d)
    cd $TMP_DIR || exit 1
    trap "rm -rf '$TMP_DIR'" EXIT
}

download() {
    file=${2:-$(basename "$1")}
    curl -sfL "$1" -o "$file" || { echo "Error: Failed to download '$2'"; exit 1; }
}

verify() {
    grep "$1" checksums.txt | sha256sum -c - \
        || { echo "Error: Checksum verification for $1 failed"; exit 1; }
}

do_install() {
    download $RELEASE_URL/$DNF_BIN
    verify $DNF_BIN

    download $RELEASE_URL/dnfd.service dnfd.service
    verify dnfd.service

    download $RELEASE_URL/config.yml config.yml
    verify config.yml

    chmod +x $DNF_BIN
    $SUDO mv $DNF_BIN $INSTALL_BIN

    $SUDO mv dnfd.service $INSTALL_SERVICE

    if [ ! -f "$INSTALL_CONFIG" ]; then
        $SUDO mkdir -p /etc/dnfd
        $SUDO mv config.yml $INSTALL_CONFIG
    fi

    if ! id "$INSTALL_USER" > /dev/null 2>&1; then
        $SUDO useradd -r -M -s /sbin/nologin $INSTALL_USER
    fi

    $SUDO systemctl daemon-reload
    $SUDO systemctl enable dnfd
}

verify_system
verify_arch

if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -sfL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name"' \
        | cut -d'"' -f4) || { echo "Error: Failed to get latest version"; exit 1; }
fi

RELEASE_URL="https://github.com/$REPO/releases/download/$VERSION/"
DNF_BIN=dnf-$OS-$ARCH

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
fi

setup_tmp
download "${RELEASE_URL}/checksums.txt" "checksums.txt"

if [ "$LOCAL_VERSION" != 0 ]; then
    echo "dnf $LOCAL_VERSION ($OS/$ARCH) is already installed"

    case $(compare_version "$LOCAL_VERSION" "$VERSION") in
        higher)
            if ask "Do you want to upgrade to dnf $VERSION ($OS/$ARCH)?"; then
                do_install
                $SUDO systemctl restart dnfd

                echo "Successfully updated dnf to $VERSION ($OS/$ARCH)."
            else
                echo "Uppgrade cancelled."
                exit 0
            fi
            ;;
        same)
            echo "Already up to date ($LOCAL_VERSION)."
            exit 0
            ;;
        lower)
            if ask "Do you want to downgrade to dnf $VERSION ($OS/$ARCH)?"; then
                do_install
                $SUDO systemctl restart dnfd

                echo "Successfully downgraded dnf to $VERSION ($OS/$ARCH)."
            else
                echo "Downgrade cancelled."
                exit 0
            fi
            ;;
    esac

    exit 0
fi

if ask "Are you sure you want to install dnf $VERSION ($OS/$ARCH)?"; then
    echo "Installing dnf $VERSION ($OS/$ARCH)"
    do_install
else
    echo "Installation cancelled."
    exit 0
fi

echo "Successfully installed dnf $VERSION ($OS/$ARCH)"
echo "The config is located at $INSTALL_CONFIG"
echo "After you have configured it, run 'systemctl start dnfd'"
