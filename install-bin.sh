#!/bin/bash
# ArchUp Installer Bootstrap Script
# Downloads and runs the ArchUp Go installer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
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

# Download latest release
info "Fetching latest release information..."
LATEST_RELEASE_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
LATEST_TAG=$(curl -s "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    error "Failed to fetch latest release tag"
fi
success "Latest version: $LATEST_TAG"

# Download binary
BINARY_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/$BINARY_NAME"
BINARY_PATH="/tmp/$BINARY_NAME"

info "Downloading installer binary..."
if ! curl -fsSL "$BINARY_URL" -o "$BINARY_PATH"; then
    error "Failed to download installer binary from $BINARY_URL"
fi
success "Binary downloaded"

# Make executable
chmod +x "$BINARY_PATH"

# Run installer
info "Launching ArchUp installer..."
echo ""
exec "$BINARY_PATH"
