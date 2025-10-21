#!/bin/bash
# Build and install archup-cli from latest GitHub release

set -e

GITHUB_REPO="bnema/archup-cli"
INSTALL_DIR="/tmp/archup-cli-build"
BINARY_PATH="/usr/bin/archup"

echo "=== Installing archup-cli ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
  echo "ERROR: Go is not installed. Cannot build archup-cli."
  exit 1
fi

# Get latest release tag from GitHub
echo "Fetching latest release from GitHub..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
  echo "WARNING: No releases found, falling back to main branch"
  DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/archive/refs/heads/main.tar.gz"
  EXTRACT_DIR="archup-cli-main"
else
  echo "Latest release: $LATEST_RELEASE"
  DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/archive/refs/tags/${LATEST_RELEASE}.tar.gz"
  # Remove 'v' prefix if present for directory name
  TAG_NAME="${LATEST_RELEASE#v}"
  EXTRACT_DIR="archup-cli-${TAG_NAME}"
fi

# Clean up any previous builds
rm -rf "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Download source
echo "Downloading source from ${DOWNLOAD_URL}..."
curl -L "$DOWNLOAD_URL" -o "${INSTALL_DIR}/archup-cli.tar.gz"

# Extract
echo "Extracting source..."
tar -xzf "${INSTALL_DIR}/archup-cli.tar.gz" -C "$INSTALL_DIR"

# Find the extracted directory
cd "${INSTALL_DIR}/${EXTRACT_DIR}"

# Build the binary
echo "Building archup-cli..."
go build -o archup .

# Install to /usr/bin
echo "Installing to ${BINARY_PATH}..."
install -m 755 archup "$BINARY_PATH"

# Cleanup
cd /
rm -rf "$INSTALL_DIR"

# Verify installation
if [ -x "$BINARY_PATH" ]; then
  echo "✓ archup-cli installed successfully"
  echo "  Location: $BINARY_PATH"
  echo "  Version: $($BINARY_PATH --version 2>/dev/null || echo 'unknown')"
else
  echo "ERROR: Installation failed"
  exit 1
fi
