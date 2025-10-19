#!/bin/bash
# Preflight phase

source "$ARCHUP_INSTALL/preflight/guards.sh"               # Interactive (gum confirm)
run_logged "$ARCHUP_INSTALL/preflight/detect-environment.sh"  # Non-interactive
source "$ARCHUP_INSTALL/preflight/begin.sh"                # Interactive (gum confirm)
