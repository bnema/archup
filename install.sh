#!/bin/bash
# archup - Fast, minimal Arch Linux auto-installer
# Main installation orchestrator

set -eEo pipefail

# Export global paths
export ARCHUP_PATH="${ARCHUP_PATH:-$HOME/.local/share/archup}"
export ARCHUP_INSTALL="$ARCHUP_PATH/install"
export ARCHUP_INSTALL_LOG_FILE="/var/log/archup-install.log"
export ARCHUP_REPO_URL="${ARCHUP_REPO_URL:-https://github.com/bnema/archup}"

# Source all helper utilities
source "$ARCHUP_INSTALL/helpers/all.sh"

# Display logo and start installation
clear_logo
start_install_log

# ============================================================
# PHASE 1: BAREBONE INSTALLER - BASIC
# ============================================================

# Preflight validation
source "$ARCHUP_INSTALL/preflight/all.sh"

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
