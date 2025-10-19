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

# Phase 0 test: Display welcome message
gum style --foreground 4 --padding "1 0 1 $PADDING_LEFT" "Welcome to archup!"
gum style --padding "0 0 0 $PADDING_LEFT" "Phase 0: Testing helper utilities..."

# Test logging
echo "Testing logging functionality..." | tee -a "$ARCHUP_INSTALL_LOG_FILE"
sleep 1

# Test gum styling
gum style --foreground 2 --padding "1 0 0 $PADDING_LEFT" "✓ Helpers loaded successfully"
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ Logging initialized"
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ Error handling ready"
gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Presentation utilities working"

# Test confirmation prompt
if gum confirm "Continue with Phase 0 test?"; then
  gum style --foreground 4 --padding "1 0 0 $PADDING_LEFT" "Great! Phase 0 infrastructure is ready."
  gum style --padding "0 0 0 $PADDING_LEFT" "Next steps:"
  gum style --padding "0 0 0 $PADDING_LEFT" "  • Phase 1: Preflight validation"
  gum style --padding "0 0 0 $PADDING_LEFT" "  • Phase 2: Partitioning & encryption"
  gum style --padding "0 0 0 $PADDING_LEFT" "  • Phase 3: Base installation"
  echo
else
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Installation cancelled by user"
  exit 0
fi

# Stop logging and cleanup
stop_install_log

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "Phase 0 complete! ✓"
