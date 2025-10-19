#!/bin/bash
# Partitioning phase

source "$ARCHUP_INSTALL/partitioning/detect-disk.sh"     # Interactive (gum confirm)
run_logged "$ARCHUP_INSTALL/partitioning/partition.sh"   # Non-interactive
run_logged "$ARCHUP_INSTALL/partitioning/format.sh"      # Non-interactive
run_logged "$ARCHUP_INSTALL/partitioning/mount.sh"       # Non-interactive
