#!/bin/bash
# ArchUp Installer Bootstrap
# Downloads the ArchUp binary from GitHub Releases and launches the installer.
#
# Usage:
#   curl -fsSL https://archup.run/install | bash
#   curl -fsSL https://archup.run/install | bash -s -- --dev
#   curl -fsSL https://archup.run/install | bash -s -- --version v0.5.0

set -e

# Parse arguments
DEV_MODE=false
SPECIFIC_VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dev)
            DEV_MODE=true
            shift
            ;;
        --version)
            SPECIFIC_VERSION="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--dev] [--version VERSION]"
            exit 1
            ;;
    esac
done

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Preflight checks
info "Checking system requirements..."

if [ ! -f /etc/arch-release ]; then
    error "Must be run from the Arch Linux ISO"
fi
success "Arch Linux detected"

if [ ! -d /sys/firmware/efi ]; then
    error "UEFI mode required (legacy BIOS not supported)"
fi
success "UEFI mode detected"

ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ]; then
    error "Only x86_64 is supported (detected: $ARCH)"
fi
success "Architecture: $ARCH"

REPO_OWNER="bnema"
REPO_NAME="archup"
BINARY_NAME="archup"

# Resolve version to download
if [ -n "$SPECIFIC_VERSION" ]; then
    LATEST_TAG="$SPECIFIC_VERSION"
    info "Using version: $LATEST_TAG"
elif [ "$DEV_MODE" = true ]; then
    warning "Dev mode: fetching latest pre-release..."
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases" \
        | grep -B 10 '"prerelease": true' \
        | grep '"tag_name":' \
        | head -n 1 \
        | sed -E 's/.*"([^"]+)".*/\1/')
    [ -z "$LATEST_TAG" ] && error "Failed to fetch latest pre-release"
    warning "Pre-release: $LATEST_TAG"
else
    info "Fetching latest stable release..."
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" \
        | grep '"tag_name":' \
        | sed -E 's/.*"([^"]+)".*/\1/')
    [ -z "$LATEST_TAG" ] && error "Failed to fetch latest release"
    success "Version: $LATEST_TAG"
fi

BINARY_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/$BINARY_NAME"
CHECKSUM_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/checksums.txt"
BINARY_PATH="/tmp/$BINARY_NAME"
CHECKSUM_PATH="/tmp/archup-checksums.txt"

info "Downloading binary..."
curl -fsSL "$BINARY_URL" -o "$BINARY_PATH" || error "Failed to download binary from $BINARY_URL"
success "Binary downloaded"

info "Verifying checksum..."
curl -fsSL "$CHECKSUM_URL" -o "$CHECKSUM_PATH" || error "Failed to download checksums"
cd /tmp
sha256sum -c "$CHECKSUM_PATH" --ignore-missing 2>/dev/null || error "Checksum verification failed"
success "Checksum verified"

chmod +x "$BINARY_PATH"

info "Launching ArchUp..."
echo ""
exec "$BINARY_PATH" install
