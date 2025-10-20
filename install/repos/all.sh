#!/bin/bash
# Repository setup phase

stop_log_output  # Stop for interactive prompts
source "$ARCHUP_INSTALL/repos/yay.sh"         # Interactive (gum confirm)
source "$ARCHUP_INSTALL/repos/chaotic.sh"     # Interactive (gum confirm)
start_log_output # Resume (will be stopped by stop_install_log)
