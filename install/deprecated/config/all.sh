#!/bin/bash
# Configuration phase

stop_log_output  # Stop for interactive prompts
source "$ARCHUP_INSTALL/config/system.sh"          # Interactive (gum input)
source "$ARCHUP_INSTALL/config/user.sh"            # Interactive (gum input)
start_log_output # Resume after prompts

run_logged "$ARCHUP_INSTALL/config/network.sh"     # Non-interactive
run_logged "$ARCHUP_INSTALL/config/zram.sh"        # Non-interactive
