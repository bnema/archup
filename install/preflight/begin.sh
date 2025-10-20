#!/bin/bash
# Begin ArchUp installation
# Display welcome message and gather initial configuration choices

# Display welcome message
gum style --foreground 4 --padding "1 0 1 $PADDING_LEFT" "Welcome to ArchUp!"
gum style --padding "0 0 0 $PADDING_LEFT" "A fast, minimal Arch Linux auto-installer"
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

# Ask for user credentials first (will be used for user account, root, and encryption)
echo
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "User Account & Security"
gum style --padding "0 0 0 $PADDING_LEFT" "Create account with sudo privileges. This password will be used for:"
gum style --padding "0 0 0 $PADDING_LEFT" "  • User login with sudo access"
gum style --padding "0 0 0 $PADDING_LEFT" "  • Disk encryption (LUKS) if enabled"
echo

# Ask for username
ARCHUP_USERNAME=$(gum input --placeholder "Enter username" \
  --prompt "Username: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ -z "$ARCHUP_USERNAME" ]; then
  gum style --foreground 1 --padding "0 0 0 $PADDING_LEFT" "Username cannot be empty!"
  exit 1
fi

# Ask for password (will be used for user, root, and LUKS encryption)
while true; do
  ARCHUP_PASSWORD=$(gum input --password --placeholder "Enter password" \
    --prompt "Password: " \
    --padding "0 0 0 $PADDING_LEFT")

  ARCHUP_PASSWORD_CONFIRM=$(gum input --password --placeholder "Confirm password" \
    --prompt "Confirm: " \
    --padding "0 0 0 $PADDING_LEFT")

  if [ "$ARCHUP_PASSWORD" = "$ARCHUP_PASSWORD_CONFIRM" ]; then
    break
  else
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "Passwords do not match. Try again."
  fi
done

export ARCHUP_USERNAME
export ARCHUP_PASSWORD

# Save to config file (passwords are never logged)
config_set "ARCHUP_USERNAME" "$ARCHUP_USERNAME"
config_set "ARCHUP_PASSWORD" "$ARCHUP_PASSWORD"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] User: $ARCHUP_USERNAME"
echo "User: $ARCHUP_USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE"

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
