#!/bin/bash
# Setup post-boot scripts and first-boot service

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Setting up first-boot service..."

# Create post-boot directory
mkdir -p /mnt/usr/local/share/archup/post-boot

# Copy logo
cp "$ARCHUP_PATH/logo.txt" /mnt/usr/local/share/archup/

# Copy all post-boot scripts
cp "$ARCHUP_INSTALL/post-boot/all.sh" /mnt/usr/local/share/archup/post-boot/
cp "$ARCHUP_INSTALL/post-boot/snapper.sh" /mnt/usr/local/share/archup/post-boot/
cp "$ARCHUP_INSTALL/post-boot/ssh-keygen.sh" /mnt/usr/local/share/archup/post-boot/
cp "$ARCHUP_INSTALL/post-boot/archup-cli.sh" /mnt/usr/local/share/archup/post-boot/

# Make all scripts executable
chmod +x /mnt/usr/local/share/archup/post-boot/*.sh

# Install systemd service (substitute email and username)
sed "s|__ARCHUP_EMAIL__|$(config_get "ARCHUP_EMAIL")|g; s|__ARCHUP_USERNAME__|$(config_get "ARCHUP_USERNAME")|g" \
  "$ARCHUP_INSTALL/post-boot/archup-first-boot.service" > /mnt/etc/systemd/system/archup-first-boot.service

# Enable the service
arch-chroot /mnt systemctl enable archup-first-boot.service >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

echo "First-boot service will run on first reboot to:" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "  - Configure snapper snapshots" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "  - Generate SSH keys" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "  - Build and install archup-cli from GitHub" >> "$ARCHUP_INSTALL_LOG_FILE"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] First-boot service configured"
