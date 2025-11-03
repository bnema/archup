#!/bin/bash
# User creation and configuration (uses credentials from preflight)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Creating user account..."

# Set root password (same as user password)
echo "root:$ARCHUP_PASSWORD" | arch-chroot /mnt chpasswd >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Create user with home directory
arch-chroot /mnt useradd -m -G wheel -s /bin/bash "$ARCHUP_USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Set user password
echo "$ARCHUP_USERNAME:$ARCHUP_PASSWORD" | arch-chroot /mnt chpasswd >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Enable sudo for wheel group
echo "%wheel ALL=(ALL:ALL) ALL" > /mnt/etc/sudoers.d/wheel
chmod 440 /mnt/etc/sudoers.d/wheel >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] User created: $ARCHUP_USERNAME"
gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Sudo enabled for wheel group"

echo "Created user: $ARCHUP_USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Enabled sudo for wheel group" >> "$ARCHUP_INSTALL_LOG_FILE"
