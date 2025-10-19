#!/bin/bash
# Generate fstab

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Generating fstab..."

# Generate fstab using UUIDs
genfstab -U /mnt >> /mnt/etc/fstab

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ fstab generated"

echo "Generated /etc/fstab" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
