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
  # Download the download helper first
  mkdir -p "$ARCHUP_INSTALL/helpers"
  curl -sL "$ARCHUP_RAW_URL/install/helpers/download.sh" -o "$ARCHUP_INSTALL/helpers/download.sh"
  chmod +x "$ARCHUP_INSTALL/helpers/download.sh"

  # Run the download script
  source "$ARCHUP_INSTALL/helpers/download.sh"
  download_archup_files
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

# Post-install (boot logo, snapper, unmount drives)
source "$ARCHUP_INSTALL/post-install/all.sh"

# Stop logging and cleanup
stop_install_log

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "ArchUp barebone installation complete!"
echo

gum style --foreground 6 --padding "0 0 0 $PADDING_LEFT" "Next steps:"
gum style --padding "0 0 0 $PADDING_LEFT" "  1. Reboot the system"
gum style --padding "0 0 0 $PADDING_LEFT" "  2. After login, run: archup wizard"
gum style --padding "0 0 1 $PADDING_LEFT" "  3. Select your compositor and packages"
echo

CHOICE=$(gum choose --header "What would you like to do?" --header.padding "0 0 0 $PADDING_LEFT" "Reboot" "Close")

if [ "$CHOICE" = "Reboot" ]; then
  gum style --padding "0 0 1 $PADDING_LEFT" "Rebooting system..."
  reboot
else
  gum style --padding "0 0 1 $PADDING_LEFT" "Installation complete. You can manually reboot when ready."
fi
