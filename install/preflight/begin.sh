#!/bin/bash
# Begin archup installation
# Display welcome message and gather initial configuration choices

# Display welcome message
gum style --foreground 4 --padding "1 0 1 $PADDING_LEFT" "Welcome to archup!"
gum style --padding "0 0 0 $PADDING_LEFT" "A fast, minimal Arch Linux auto-installer"
echo

# Confirm installation start
if ! gum confirm "Ready to begin installation?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Installation cancelled by user"
  exit 0
fi

# Ask for bootloader preference (Phase 1 requirement)
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Bootloader Selection"
gum style --padding "0 0 0 $PADDING_LEFT" "archup supports modern bootloaders for UEFI systems:"
echo

ARCHUP_BOOTLOADER=$(gum choose \
  "systemd-boot" \
  "limine" \
  --header "Select bootloader:" \
  --height 4 \
  --padding "0 0 0 $PADDING_LEFT")

export ARCHUP_BOOTLOADER
echo "Selected bootloader: $ARCHUP_BOOTLOADER" | tee -a "$ARCHUP_INSTALL_LOG_FILE"

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "✓ Configuration saved"
