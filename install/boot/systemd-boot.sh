#!/bin/bash
# Install and configure systemd-boot

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing systemd-boot..."

# Install systemd-boot
arch-chroot /mnt bootctl install

# Determine root device UUID
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  # For encrypted setups, we need the root partition UUID (before LUKS)
  ROOT_UUID=$(blkid -s UUID -o value "$ARCHUP_ROOT_PART")

  # Update mkinitcpio.conf to add encrypt hook
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring initramfs for encryption..."

  # Add encrypt hook before filesystems hook
  sed -i 's/^HOOKS=.*/HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block encrypt filesystems fsck)/' /mnt/etc/mkinitcpio.conf

  # Regenerate initramfs
  arch-chroot /mnt mkinitcpio -P

  echo "Updated mkinitcpio.conf with encrypt hook" | tee -a "$ARCHUP_INSTALL_LOG_FILE"

  # Kernel parameters for encrypted boot
  KERNEL_PARAMS="cryptdevice=UUID=$ROOT_UUID:cryptroot root=/dev/mapper/cryptroot rootflags=subvol=@ rw"
else
  # For non-encrypted setups, use the root partition UUID directly
  ROOT_UUID=$(blkid -s UUID -o value "$ARCHUP_ROOT_PART")
  KERNEL_PARAMS="root=UUID=$ROOT_UUID rootflags=subvol=@ rw"
fi

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
options $KERNEL_PARAMS
EOF

# Create fallback boot entry
cat > /mnt/boot/loader/entries/arch-fallback.conf <<EOF
title   Arch Linux (fallback)
linux   /vmlinuz-linux
initrd  /initramfs-linux-fallback.img
options $KERNEL_PARAMS
EOF

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ systemd-boot installed and configured"

if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  echo "Installed systemd-boot with encrypted root (UUID: $ROOT_UUID)" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
else
  echo "Installed systemd-boot with root UUID: $ROOT_UUID" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
fi
