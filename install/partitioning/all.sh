#!/bin/bash
# Partitioning phase orchestrator

run_logged "$ARCHUP_INSTALL/partitioning/detect-disk.sh"
run_logged "$ARCHUP_INSTALL/partitioning/partition.sh"
run_logged "$ARCHUP_INSTALL/partitioning/format.sh"
run_logged "$ARCHUP_INSTALL/partitioning/mount.sh"
