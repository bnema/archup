#!/bin/bash
# Auto-partition disk with GPT layout (EFI + root)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Creating partitions..."

# Wipe disk and create GPT partition table
wipefs -af "$ARCHUP_DISK"
sgdisk --zap-all "$ARCHUP_DISK"

# Create partitions
# 1. EFI partition (512MB)
# 2. Root partition (remaining space)
sgdisk --clear \
  --new=1:0:+512M --typecode=1:ef00 --change-name=1:"EFI" \
  --new=2:0:0 --typecode=2:8300 --change-name=2:"ROOT" \
  "$ARCHUP_DISK"

# Inform kernel of partition changes
partprobe "$ARCHUP_DISK"
sleep 1

# Set partition variables
if [[ "$ARCHUP_DISK" =~ nvme ]]; then
  export ARCHUP_EFI_PART="${ARCHUP_DISK}p1"
  export ARCHUP_ROOT_PART="${ARCHUP_DISK}p2"
else
  export ARCHUP_EFI_PART="${ARCHUP_DISK}1"
  export ARCHUP_ROOT_PART="${ARCHUP_DISK}2"
fi

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ Partitions created"
gum style --padding "0 0 0 $PADDING_LEFT" "  EFI: $ARCHUP_EFI_PART (512MB)"
gum style --padding "0 0 1 $PADDING_LEFT" "  Root: $ARCHUP_ROOT_PART"

echo "EFI partition: $ARCHUP_EFI_PART" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Root partition: $ARCHUP_ROOT_PART" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
