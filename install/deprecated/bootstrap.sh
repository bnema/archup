#!/bin/bash
# Bootstrap: Install essential dependencies before running preflight
# This script cannot use gum since it installs gum!
# Must be run with internet connectivity (after wifi setup in ISO)

set -e

# Clean up old log file to ensure fresh logs for this attempt
if [ -f "$ARCHUP_INSTALL_LOG_FILE" ]; then
  rm -f "$ARCHUP_INSTALL_LOG_FILE"
  echo "Cleared previous installation log"
fi

echo "=== ArchUp Bootstrap ==="
echo "Installing essential dependencies..."

# Fix TERM for better gum compatibility in VMs
if [ "$TERM" = "linux" ] || [ -z "$TERM" ]; then
    export TERM=xterm-256color
    echo "Set TERM=xterm-256color for better TUI compatibility"
fi

# Sync package database (required on fresh ISO boot)
echo "Syncing package database..."
pacman -Sy --noconfirm

# Install gum for TUI
if ! command -v gum &>/dev/null; then
    echo "Installing gum for interactive TUI..."
    pacman -S --needed --noconfirm gum
    echo "[OK] gum installed"
else
    echo "[OK] gum already available"
fi

# Install jq for JSON parsing (presets)
if ! command -v jq &>/dev/null; then
    echo "Installing jq for preset JSON parsing..."
    pacman -S --needed --noconfirm jq
    echo "[OK] jq installed"
else
    echo "[OK] jq already available"
fi

# Any other critical dependencies that preflight needs can go here
# Example: pacman -S --needed --noconfirm other-tool

echo "[OK] Bootstrap complete!"
echo
