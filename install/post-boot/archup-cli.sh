#!/bin/bash
# Build and install archup-cli using go install

set -e

GITHUB_REPO="bnema/archup-cli"

echo "=== Installing archup-cli ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
  echo "ERROR: Go is not installed. Cannot install archup-cli."
  exit 1
fi

# Get latest release tag from GitHub
echo "Fetching latest release from GitHub..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
  echo "WARNING: No releases found, installing from main branch"
  VERSION="@latest"
else
  echo "Latest release: $LATEST_RELEASE"
  VERSION="@${LATEST_RELEASE}"
fi

# Install using go install (installs to $GOBIN or $GOPATH/bin or ~/go/bin)
echo "Installing archup-cli ${VERSION}..."
go install "github.com/${GITHUB_REPO}${VERSION}"

# Determine where the binary was installed
GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
BINARY_PATH="${GOBIN}/archup-cli"

# Verify installation
if [ -x "$BINARY_PATH" ]; then
  echo "✓ archup-cli installed successfully"
  echo "  Location: $BINARY_PATH"
  echo "  Version: $($BINARY_PATH --version 2>/dev/null || echo 'unknown')"
  echo ""
  echo "To update in the future, run:"
  echo "  go install github.com/${GITHUB_REPO}@latest"
else
  echo "ERROR: Installation failed"
  exit 1
fi
