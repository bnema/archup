#!/bin/bash
# Preflight phase

source "$ARCHUP_INSTALL/preflight/guards.sh"               # Interactive (gum confirm)
source "$ARCHUP_INSTALL/preflight/detect-environment.sh"   # Exports ARCHUP_KEYMAP
source "$ARCHUP_INSTALL/preflight/begin.sh"                # Interactive (gum confirm) + exports ARCHUP_BOOTLOADER, ARCHUP_ENCRYPTION
