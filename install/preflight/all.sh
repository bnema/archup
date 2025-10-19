#!/bin/bash
# Preflight phase orchestrator
# Sources all preflight scripts in order

# Run guards (system validation)
source "$ARCHUP_INSTALL/preflight/guards.sh"

# Detect environment (keyboard layout, WiFi)
run_logged "$ARCHUP_INSTALL/preflight/detect-environment.sh"

# Run begin (welcome and initial config)
run_logged "$ARCHUP_INSTALL/preflight/begin.sh"
