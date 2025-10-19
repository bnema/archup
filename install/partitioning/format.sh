#!/bin/bash
# Format partitions with btrfs and optional LUKS encryption

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Formatting partitions..."

# Format EFI partition as FAT32
mkfs.fat -F32 -n EFI "$ARCHUP_EFI_PART" 2>&1
echo "Formatted $ARCHUP_EFI_PART as FAT32" >> "$ARCHUP_INSTALL_LOG_FILE"

# Handle root partition - with or without encryption
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Setting up LUKS encryption..."

  # Prompt for encryption password
  while true; do
    LUKS_PASSWORD=$(gum input --password --placeholder "Enter encryption password" --padding "0 0 0 $PADDING_LEFT")
    LUKS_PASSWORD_CONFIRM=$(gum input --password --placeholder "Confirm encryption password" --padding "0 0 0 $PADDING_LEFT")

    if [ "$LUKS_PASSWORD" = "$LUKS_PASSWORD_CONFIRM" ]; then
      break
    else
      gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "Passwords do not match. Try again."
    fi
  done

  # Setup LUKS with Argon2id and 2000ms iteration time
  echo "$LUKS_PASSWORD" | cryptsetup luksFormat \
    --type luks2 \
    --pbkdf argon2id \
    --iter-time 2000 \
    --label ARCHUP_LUKS \
    "$ARCHUP_ROOT_PART"

  # Open the encrypted container
  echo "$LUKS_PASSWORD" | cryptsetup open "$ARCHUP_ROOT_PART" cryptroot

  # Export the mapped device path
  export ARCHUP_CRYPT_ROOT="/dev/mapper/cryptroot"

  # Format the encrypted container with btrfs
  mkfs.btrfs -f -L ROOT "$ARCHUP_CRYPT_ROOT"

  echo "Created LUKS container on $ARCHUP_ROOT_PART" >> "$ARCHUP_INSTALL_LOG_FILE"
  echo "Formatted $ARCHUP_CRYPT_ROOT as btrfs" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Mount temporarily to create subvolumes
  mount "$ARCHUP_CRYPT_ROOT" /mnt

else
  # No encryption - format directly with btrfs
  mkfs.btrfs -f -L ROOT "$ARCHUP_ROOT_PART"
  echo "Formatted $ARCHUP_ROOT_PART as btrfs" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Mount temporarily to create subvolumes
  mount "$ARCHUP_ROOT_PART" /mnt
fi

# Create btrfs subvolumes
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Creating btrfs subvolumes..."

btrfs subvolume create /mnt/@
btrfs subvolume create /mnt/@home

echo "Created btrfs subvolume: @" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Created btrfs subvolume: @home" >> "$ARCHUP_INSTALL_LOG_FILE"

# Unmount temporary mount
umount /mnt

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Partitions formatted and subvolumes created"
echo "Partitions formatted and subvolumes created successfully" >> "$ARCHUP_INSTALL_LOG_FILE"
