#!/bin/bash
# Base installation phase orchestrator

run_logged "$ARCHUP_INSTALL/base/pacstrap.sh"
run_logged "$ARCHUP_INSTALL/base/fstab.sh"
