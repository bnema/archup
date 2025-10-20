#!/bin/bash
# Base installation phase

stop_log_output  # Stop for interactive prompt
source "$ARCHUP_INSTALL/base/kernel.sh"           # Interactive (gum choose)
start_log_output # Resume after prompt

run_logged "$ARCHUP_INSTALL/base/pacstrap.sh"     # Non-interactive
run_logged "$ARCHUP_INSTALL/base/fstab.sh"        # Non-interactive
