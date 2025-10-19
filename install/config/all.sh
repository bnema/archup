#!/bin/bash
# Configuration phase orchestrator

run_logged "$ARCHUP_INSTALL/config/system.sh"
run_logged "$ARCHUP_INSTALL/config/user.sh"
run_logged "$ARCHUP_INSTALL/config/network.sh"
