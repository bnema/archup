#!/bin/bash
# Begin ArchUp installation
# Display welcome message and gather initial configuration choices

# Display welcome message
gum style --foreground 4 --padding "1 0 1 $PADDING_LEFT" "Welcome to ArchUp!"
gum style --padding "0 0 0 $PADDING_LEFT" "A minimal, slightly opinionated Arch Linux installer."
echo

# Confirm installation start
if ! gum confirm "Ready to begin installation?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Installation cancelled by user"
  echo "Installation cancelled by user" >> "$ARCHUP_INSTALL_LOG_FILE"
  exit 0
fi

# ArchUp uses Limine bootloader exclusively
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Bootloader: Limine"
gum style --padding "0 0 1 $PADDING_LEFT" "ArchUp uses Limine for superior btrfs and snapshot support"

export ARCHUP_BOOTLOADER="limine"
config_set "ARCHUP_BOOTLOADER" "limine"
echo "Bootloader: limine" >> "$ARCHUP_INSTALL_LOG_FILE"

# Gather user identity (username, password, email, hostname, timezone)
echo
source "$ARCHUP_INSTALL/preflight/identify.sh"

# Ask for encryption preference
echo
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Disk Encryption"
gum style --padding "0 0 0 $PADDING_LEFT" "LUKS encryption protects your data with Argon2id (2000ms iteration)"
echo

if gum confirm "Enable full-disk encryption?" --padding "0 0 1 $PADDING_LEFT"; then
  export ARCHUP_ENCRYPTION="enabled"
  config_set "ARCHUP_ENCRYPTION" "enabled"
  echo "Encryption: enabled" >> "$ARCHUP_INSTALL_LOG_FILE"
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "Note: Using account password for disk encryption"
else
  export ARCHUP_ENCRYPTION="disabled"
  config_set "ARCHUP_ENCRYPTION" "disabled"
  echo "Encryption: disabled" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Configuration saved"
echo "Configuration saved" >> "$ARCHUP_INSTALL_LOG_FILE"
