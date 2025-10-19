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
  echo "Installation cancelled by user" >> "$ARCHUP_INSTALL_LOG_FILE"
  exit 0
fi

# archup uses Limine bootloader exclusively
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Bootloader: Limine"
gum style --padding "0 0 1 $PADDING_LEFT" "archup uses Limine for superior btrfs and snapshot support"

export ARCHUP_BOOTLOADER="limine"
echo "Bootloader: limine" >> "$ARCHUP_INSTALL_LOG_FILE"

# Ask for encryption preference (Phase 2)
echo
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Disk Encryption"
gum style --padding "0 0 0 $PADDING_LEFT" "LUKS encryption protects your data with Argon2id (2000ms iteration)"
echo

if gum confirm "Enable full-disk encryption?" --padding "0 0 1 $PADDING_LEFT"; then
  export ARCHUP_ENCRYPTION="enabled"
  echo "Encryption: enabled" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  export ARCHUP_ENCRYPTION="disabled"
  echo "Encryption: disabled" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Configuration saved"
echo "Configuration saved" >> "$ARCHUP_INSTALL_LOG_FILE"
