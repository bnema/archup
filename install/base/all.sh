#!/bin/bash
# Base installation phase

source "$ARCHUP_INSTALL/base/kernel.sh"           # Interactive (gum choose)
run_logged "$ARCHUP_INSTALL/base/pacstrap.sh"     # Non-interactive
run_logged "$ARCHUP_INSTALL/base/fstab.sh"        # Non-interactive
