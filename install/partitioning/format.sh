#!/bin/bash
# Format partitions with btrfs and optional LUKS encryption

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Formatting partitions..."

# Wipe any existing signatures on EFI partition
wipefs -af "$ARCHUP_EFI_PART" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Format EFI partition as FAT32
mkfs.fat -F32 -n EFI "$ARCHUP_EFI_PART" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
echo "Formatted $ARCHUP_EFI_PART as FAT32" >> "$ARCHUP_INSTALL_LOG_FILE"

# Handle root partition - with or without encryption
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Setting up LUKS encryption..."

  # Wipe any existing signatures on root partition to avoid conflicts
  wipefs -af "$ARCHUP_ROOT_PART" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  # Setup LUKS with Argon2id and 2000ms iteration time (using user password)
  echo "$ARCHUP_PASSWORD" | cryptsetup luksFormat \
    --type luks2 \
    --batch-mode \
    --pbkdf argon2id \
    --iter-time 2000 \
    --label ARCHUP_LUKS \
    "$ARCHUP_ROOT_PART" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  echo "LUKS container created" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Open the encrypted container
  echo "$ARCHUP_PASSWORD" | cryptsetup open "$ARCHUP_ROOT_PART" cryptroot >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  echo "LUKS container opened" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Export the mapped device path
  export ARCHUP_CRYPT_ROOT="/dev/mapper/cryptroot"

  # Format the encrypted container with btrfs
  mkfs.btrfs -f -L ROOT "$ARCHUP_CRYPT_ROOT" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  echo "Created LUKS container on $ARCHUP_ROOT_PART" >> "$ARCHUP_INSTALL_LOG_FILE"
  echo "Formatted $ARCHUP_CRYPT_ROOT as btrfs" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Mount temporarily to create subvolumes
  mount "$ARCHUP_CRYPT_ROOT" /mnt >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

else
  # No encryption - format directly with btrfs
  mkfs.btrfs -f -L ROOT "$ARCHUP_ROOT_PART" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
  echo "Formatted $ARCHUP_ROOT_PART as btrfs" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Mount temporarily to create subvolumes
  mount "$ARCHUP_ROOT_PART" /mnt >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
fi

# Create btrfs subvolumes
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Creating btrfs subvolumes..."

btrfs subvolume create /mnt/@ >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
btrfs subvolume create /mnt/@home >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

echo "Created btrfs subvolume: @" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Created btrfs subvolume: @home" >> "$ARCHUP_INSTALL_LOG_FILE"

# Unmount temporary mount
umount /mnt >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Partitions formatted and subvolumes created"
echo "Partitions formatted and subvolumes created successfully" >> "$ARCHUP_INSTALL_LOG_FILE"
