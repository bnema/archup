#!/bin/bash
# Install and configure systemd-boot

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing systemd-boot..."

# Install systemd-boot
arch-chroot /mnt bootctl install

# Get root partition UUID
ROOT_UUID=$(blkid -s UUID -o value "$ARCHUP_ROOT_PART")

# Create loader configuration
cat > /mnt/boot/loader/loader.conf <<EOF
default arch.conf
timeout 3
console-mode max
editor no
EOF

# Create boot entry
cat > /mnt/boot/loader/entries/arch.conf <<EOF
title   Arch Linux
linux   /vmlinuz-linux
initrd  /initramfs-linux.img
options root=UUID=$ROOT_UUID rw
EOF

# Create fallback boot entry
cat > /mnt/boot/loader/entries/arch-fallback.conf <<EOF
title   Arch Linux (fallback)
linux   /vmlinuz-linux
initrd  /initramfs-linux-fallback.img
options root=UUID=$ROOT_UUID rw
EOF

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ systemd-boot installed and configured"

echo "Installed systemd-boot with root UUID: $ROOT_UUID" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
