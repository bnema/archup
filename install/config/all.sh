#!/bin/bash
# Configuration phase

source "$ARCHUP_INSTALL/config/system.sh"          # Interactive (gum input)
source "$ARCHUP_INSTALL/config/user.sh"            # Interactive (gum input)
run_logged "$ARCHUP_INSTALL/config/network.sh"     # Non-interactive
