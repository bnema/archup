#!/bin/bash
# Base installation phase

stop_log_output  # Stop for interactive prompt
source "$ARCHUP_INSTALL/base/kernel.sh"           # Interactive (gum choose)
start_log_output # Resume after prompt

run_logged "$ARCHUP_INSTALL/base/cachyos-repo.sh" # Configure CachyOS repo if needed
run_logged "$ARCHUP_INSTALL/base/pacstrap.sh"     # Non-interactive (essential packages only)
run_logged "$ARCHUP_INSTALL/base/fstab.sh"        # Non-interactive
