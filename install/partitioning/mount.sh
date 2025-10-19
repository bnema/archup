#!/bin/bash
# Mount partitions to /mnt

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Mounting partitions..."

# Mount root partition
mount "$ARCHUP_ROOT_PART" /mnt

# Create and mount EFI directory
mkdir -p /mnt/boot
mount "$ARCHUP_EFI_PART" /mnt/boot

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Partitions mounted"

echo "Mounted $ARCHUP_ROOT_PART to /mnt" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Mounted $ARCHUP_EFI_PART to /mnt/boot" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
