#!/bin/bash
# Format partitions (ext4 for Phase 1, btrfs comes in Phase 2)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Formatting partitions..."

# Format EFI partition as FAT32
mkfs.fat -F32 -n EFI "$ARCHUP_EFI_PART"

# Format root partition as ext4
mkfs.ext4 -F -L ROOT "$ARCHUP_ROOT_PART"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Partitions formatted"

echo "Formatted $ARCHUP_EFI_PART as FAT32" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Formatted $ARCHUP_ROOT_PART as ext4" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
