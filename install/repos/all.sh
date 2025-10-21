#!/bin/bash
# Repository setup phase

stop_log_output  # Stop for interactive prompts

source "$ARCHUP_INSTALL/repos/chaotic.sh"     # Interactive (gum confirm) - MUST be before AUR
source "$ARCHUP_INSTALL/repos/aur.sh"         # Interactive (gum confirm + choose)

start_log_output # Resume (will be stopped by stop_install_log)
