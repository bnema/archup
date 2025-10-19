#!/bin/bash
# Bootstrap: Install essential dependencies before running preflight
# This script cannot use gum since it installs gum!
# Must be run with internet connectivity (after wifi setup in ISO)

set -e

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
    echo "✓ gum installed"
else
    echo "✓ gum already available"
fi

# Any other critical dependencies that preflight needs can go here
# Example: pacman -S --needed --noconfirm other-tool

echo "✓ Bootstrap complete!"
echo
