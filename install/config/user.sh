#!/bin/bash
# User creation and configuration

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "User Account Setup"
echo

# Ask for username
ARCHUP_USERNAME=$(gum input --placeholder "Enter username" \
  --prompt "Username: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ -z "$ARCHUP_USERNAME" ]; then
  gum style --foreground 1 --padding "0 0 0 $PADDING_LEFT" "Username cannot be empty!"
  exit 1
fi

# Ask for password
ARCHUP_PASSWORD=$(gum input --password --placeholder "Enter password" \
  --prompt "Password: " \
  --padding "0 0 0 $PADDING_LEFT")

ARCHUP_PASSWORD_CONFIRM=$(gum input --password --placeholder "Confirm password" \
  --prompt "Confirm: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ "$ARCHUP_PASSWORD" != "$ARCHUP_PASSWORD_CONFIRM" ]; then
  gum style --foreground 1 --padding "0 0 0 $PADDING_LEFT" "Passwords do not match!"
  exit 1
fi

echo

# Set root password (same as user password for simplicity in Phase 1)
echo "root:$ARCHUP_PASSWORD" | arch-chroot /mnt chpasswd

# Create user with home directory
arch-chroot /mnt useradd -m -G wheel -s /bin/bash "$ARCHUP_USERNAME"

# Set user password
echo "$ARCHUP_USERNAME:$ARCHUP_PASSWORD" | arch-chroot /mnt chpasswd

# Enable sudo for wheel group
echo "%wheel ALL=(ALL:ALL) ALL" > /mnt/etc/sudoers.d/wheel
chmod 440 /mnt/etc/sudoers.d/wheel

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ User created: $ARCHUP_USERNAME"
gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Sudo enabled for wheel group"

echo "Created user: $ARCHUP_USERNAME" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Enabled sudo for wheel group" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
