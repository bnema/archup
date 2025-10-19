#!/bin/bash
# Mount btrfs subvolumes to /mnt

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Mounting partitions..."

# Determine which device to mount (encrypted or not)
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  MOUNT_DEVICE="$ARCHUP_CRYPT_ROOT"
else
  MOUNT_DEVICE="$ARCHUP_ROOT_PART"
fi

# Mount @ subvolume as root with optimal btrfs options
mount -o noatime,compress=zstd,subvol=@ "$MOUNT_DEVICE" /mnt

echo "Mounted @ subvolume to /mnt" >> "$ARCHUP_INSTALL_LOG_FILE"

# Create home directory
mkdir -p /mnt/home

# Mount @home subvolume
mount -o noatime,compress=zstd,subvol=@home "$MOUNT_DEVICE" /mnt/home

echo "Mounted @home subvolume to /mnt/home" >> "$ARCHUP_INSTALL_LOG_FILE"

# Create and mount EFI directory
mkdir -p /mnt/boot
mount "$ARCHUP_EFI_PART" /mnt/boot

echo "Mounted $ARCHUP_EFI_PART to /mnt/boot" >> "$ARCHUP_INSTALL_LOG_FILE"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Partitions mounted"
echo "Partitions mounted successfully" >> "$ARCHUP_INSTALL_LOG_FILE"
