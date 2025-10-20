#!/bin/bash
# Partitioning phase

stop_log_output  # Stop for interactive prompts and partition display
source "$ARCHUP_INSTALL/partitioning/detect-disk.sh"     # Interactive (gum confirm)
source "$ARCHUP_INSTALL/partitioning/partition.sh"       # Exports ARCHUP_EFI_PART, ARCHUP_ROOT_PART (has gum output)
source "$ARCHUP_INSTALL/partitioning/format.sh"          # Interactive if encryption enabled + exports ARCHUP_CRYPT_ROOT
source "$ARCHUP_INSTALL/partitioning/mount.sh"           # Uses exported partition variables (has gum output)
start_log_output # Resume after all partitioning is done
