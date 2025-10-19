#!/bin/bash
# Detect and select installation disk

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Disk Selection"
echo

# List available disks
mapfile -t disks < <(lsblk -dno NAME,SIZE,TYPE | grep disk | awk '{print "/dev/" $1 " (" $2 ")"}')

if [ ${#disks[@]} -eq 0 ]; then
  gum style --foreground 1 --padding "0 0 0 $PADDING_LEFT" "No disks found!"
  echo "ERROR: No disks found" >> "$ARCHUP_INSTALL_LOG_FILE"
  exit 1
fi

# Let user select disk
ARCHUP_DISK=$(gum choose "${disks[@]}" \
  --header "Select installation disk (ALL DATA WILL BE ERASED):" \
  --height $((${#disks[@]} + 3)) \
  --padding "0 0 0 $PADDING_LEFT" | awk '{print $1}')

export ARCHUP_DISK

# Display selected disk info
gum style --padding "1 0 0 $PADDING_LEFT" "Selected disk: $ARCHUP_DISK"
lsblk "$ARCHUP_DISK" | while IFS= read -r line; do
  gum style --padding "0 0 0 $PADDING_LEFT" "  $line"
done
echo

# Final confirmation
if ! gum confirm "WARNING: All data on $ARCHUP_DISK will be permanently erased. Continue?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Installation cancelled by user"
  echo "Installation cancelled by user" >> "$ARCHUP_INSTALL_LOG_FILE"
  exit 0
fi

echo "Disk selected: $ARCHUP_DISK" >> "$ARCHUP_INSTALL_LOG_FILE"
