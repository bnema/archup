#!/bin/bash
# ArchUp Installer Bootstrap Script - Dev/Pre-release Version
# Downloads and runs the latest ArchUp Go installer pre-release
#
# Usage:
#   curl -fsSL https://archup.run/install/bin/dev | bash
#
# Or locally:
#   ./install-bin-dev.sh           # Install latest pre-release
#   ./install-bin-dev.sh --version v0.4.0-dev  # Install specific version

set -e

# Default to dev mode
DEV_MODE=true
SPECIFIC_VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            SPECIFIC_VERSION="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--version VERSION]"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Check if running on Arch Linux
info "Checking system requirements..."
if [ ! -f /etc/arch-release ]; then
    error "This installer must be run on Arch Linux ISO"
fi
success "Running on Arch Linux"

# Check for UEFI mode
if [ ! -d /sys/firmware/efi ]; then
    error "UEFI mode required (legacy BIOS not supported)"
fi
success "UEFI mode detected"

# Check architecture
ARCH=$(uname -m)
if [ "$ARCH" != "x86_64" ]; then
    error "Only x86_64 architecture is supported (detected: $ARCH)"
fi
success "Architecture: $ARCH"

# GitHub repository details
REPO_OWNER="bnema"
REPO_NAME="archup"
BINARY_NAME="archup-installer"

# Determine which version to download
if [ -n "$SPECIFIC_VERSION" ]; then
    # User specified a version
    LATEST_TAG="$SPECIFIC_VERSION"
    info "Using specific version: $LATEST_TAG"
else
    # Get latest pre-release (including dev/rc builds)
    warning "Dev mode: fetching latest pre-release..."
    RELEASES_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases"
    # Filter for releases where "prerelease": true, then get the first tag_name
    LATEST_TAG=$(curl -s "$RELEASES_URL" | grep -B 10 '"prerelease": true' | grep '"tag_name":' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_TAG" ]; then
        error "Failed to fetch latest pre-release"
    fi
    warning "Latest pre-release: $LATEST_TAG"
fi

# Download binary
BINARY_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/$BINARY_NAME"
BINARY_PATH="/tmp/$BINARY_NAME"
CHECKSUM_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/checksums.txt"
CHECKSUM_PATH="/tmp/checksums.txt"

info "Downloading installer binary..."
if ! curl -fsSL "$BINARY_URL" -o "$BINARY_PATH"; then
    error "Failed to download installer binary from $BINARY_URL"
fi
success "Binary downloaded"

# Download and verify checksum
info "Downloading checksums..."
if ! curl -fsSL "$CHECKSUM_URL" -o "$CHECKSUM_PATH"; then
    error "Failed to download checksums from $CHECKSUM_URL"
fi
success "Checksums downloaded"

info "Verifying checksum..."
cd /tmp
if ! sha256sum -c "$CHECKSUM_PATH" --ignore-missing 2>/dev/null; then
    error "Checksum verification failed! Binary may be corrupted or tampered with."
fi
success "Checksum verified"

# Make executable
chmod +x "$BINARY_PATH"

# Run installer
info "Launching ArchUp installer..."
echo ""
exec "$BINARY_PATH"
