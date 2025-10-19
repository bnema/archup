#!/bin/bash
# ArchUp - Fast, minimal Arch Linux auto-installer
# Main installation orchestrator

set -eo pipefail

# Export global paths
export ARCHUP_PATH="${ARCHUP_PATH:-$HOME/.local/share/archup}"
export ARCHUP_INSTALL="$ARCHUP_PATH/install"
export ARCHUP_INSTALL_LOG_FILE="/var/log/archup-install.log"
export ARCHUP_INSTALL_CONFIG="/var/log/archup-install.conf"
export ARCHUP_REPO_URL="${ARCHUP_REPO_URL:-https://github.com/bnema/archup}"
export ARCHUP_RAW_URL="${ARCHUP_RAW_URL:-https://raw.githubusercontent.com/bnema/archup/dev}"

# Handle command-line flags
if [ "$1" = "--cleanup" ]; then
  echo "Running cleanup before installation..."
  source "$ARCHUP_INSTALL/helpers/cleanup.sh" default
fi

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  cat << 'HELP'
Usage: install.sh [OPTIONS]

OPTIONS:
  --cleanup     Cleanup before running installation
  -h, --help    Show this help message

EXAMPLES:
  install.sh                 # Run normal installation
  install.sh --cleanup       # Cleanup then run installation
  ./cleanup.sh default       # Manual cleanup only
  ./cleanup.sh full          # Full cleanup including /mnt wipe
  ./cleanup.sh diagnostic    # Show current system state
HELP
  exit 0
fi

# Download installer files if not present (for curl-based installation)
if [ ! -d "$ARCHUP_INSTALL" ]; then
  echo "=== Downloading ArchUp installer files ==="
  mkdir -p "$ARCHUP_INSTALL"

  # Download all required files using curl
  GITHUB_RAW="$ARCHUP_RAW_URL"

  # Create directory structure
  mkdir -p "$ARCHUP_INSTALL"/{helpers,preflight,partitioning,base,config,boot,repos,post-install,presets}

  echo "Downloading core files..."
  curl -sL "$GITHUB_RAW/install/bootstrap.sh" -o "$ARCHUP_INSTALL/bootstrap.sh"
  curl -sL "$GITHUB_RAW/logo.txt" -o "$ARCHUP_PATH/logo.txt"

  echo "Downloading helpers..."
  for file in all.sh config.sh logging.sh errors.sh presentation.sh chroot.sh cleanup.sh; do
    curl -sL "$GITHUB_RAW/install/helpers/$file" -o "$ARCHUP_INSTALL/helpers/$file"
  done

  echo "Downloading preflight..."
  for file in all.sh guards.sh begin.sh identify.sh detect-environment.sh; do
    curl -sL "$GITHUB_RAW/install/preflight/$file" -o "$ARCHUP_INSTALL/preflight/$file"
  done

  echo "Downloading partitioning..."
  for file in all.sh detect-disk.sh partition.sh format.sh mount.sh; do
    curl -sL "$GITHUB_RAW/install/partitioning/$file" -o "$ARCHUP_INSTALL/partitioning/$file"
  done

  echo "Downloading base..."
  for file in all.sh kernel.sh pacstrap.sh fstab.sh; do
    curl -sL "$GITHUB_RAW/install/base/$file" -o "$ARCHUP_INSTALL/base/$file"
  done

  echo "Downloading config..."
  for file in all.sh system.sh user.sh network.sh; do
    curl -sL "$GITHUB_RAW/install/config/$file" -o "$ARCHUP_INSTALL/config/$file"
  done

  echo "Downloading boot..."
  for file in all.sh limine.sh; do
    curl -sL "$GITHUB_RAW/install/boot/$file" -o "$ARCHUP_INSTALL/boot/$file"
  done

  echo "Downloading repos..."
  for file in all.sh yay.sh chaotic.sh; do
    curl -sL "$GITHUB_RAW/install/repos/$file" -o "$ARCHUP_INSTALL/repos/$file"
  done

  echo "Downloading post-install..."
  for file in all.sh boot-logo.sh snapper.sh verify.sh unmount.sh; do
    curl -sL "$GITHUB_RAW/install/post-install/$file" -o "$ARCHUP_INSTALL/post-install/$file"
  done

  echo "Downloading post-boot..."
  mkdir -p "$ARCHUP_INSTALL/post-boot"
  for file in all.sh snapper.sh ssh-keygen.sh archup-first-boot.service; do
    curl -sL "$GITHUB_RAW/install/post-boot/$file" -o "$ARCHUP_INSTALL/post-boot/$file"
  done

  echo "Downloading presets..."
  curl -sL "$GITHUB_RAW/install/presets/barebone.packages" -o "$ARCHUP_INSTALL/presets/barebone.packages"

  echo "[OK] All files downloaded successfully"
  echo ""
fi

# Bootstrap: Install gum and essential dependencies (plain text, no gum usage)
source "$ARCHUP_INSTALL/bootstrap.sh"

# Source all helper utilities (now safe to use gum)
source "$ARCHUP_INSTALL/helpers/all.sh"

# Display logo
clear_logo

# ============================================================
# PHASE 1: BAREBONE INSTALLER - BASIC
# ============================================================

# Initialize log file
start_install_log

# Preflight validation (interactive prompts, no log monitor running)
source "$ARCHUP_INSTALL/preflight/all.sh"

# Start log monitor ONCE - it will run continuously during all work phases
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing..."
echo
start_log_output

# Partitioning (auto GPT, ext4, no encryption)
source "$ARCHUP_INSTALL/partitioning/all.sh"

# Base system installation
source "$ARCHUP_INSTALL/base/all.sh"

# System configuration
source "$ARCHUP_INSTALL/config/all.sh"

# Bootloader installation
source "$ARCHUP_INSTALL/boot/all.sh"

# Repository setup (AUR + Chaotic)
source "$ARCHUP_INSTALL/repos/all.sh"

# Post-install (boot logo, unmount drives)
source "$ARCHUP_INSTALL/post-install/all.sh"

# ============================================================
# FUTURE PHASES (TO BE IMPLEMENTED)
# ============================================================
# Phase 6: Barebone preset testing
# Phase 7: Default preset (GUI)

# Stop logging and cleanup
stop_install_log

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "ArchUp installation complete!"
gum style --padding "0 0 0 $PADDING_LEFT" "You can now reboot into your new Arch Linux system."
echo
