#!/bin/bash
# Install and configure Limine bootloader

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing Limine bootloader..."

# Determine root device UUID and prepare kernel parameters
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  # For encrypted setups, we need the root partition UUID (before LUKS)
  ROOT_UUID=$(blkid -s UUID -o value "$ARCHUP_ROOT_PART")

  # Update mkinitcpio.conf to add encrypt hook
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring initramfs for encryption..."

  # Add encrypt hook before filesystems hook
  sed -i 's/^HOOKS=.*/HOOKS=(base udev autodetect microcode modconf kms keyboard keymap consolefont block encrypt filesystems fsck)/' /mnt/etc/mkinitcpio.conf

  # Regenerate initramfs
  arch-chroot /mnt mkinitcpio -P >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  echo "Updated mkinitcpio.conf with encrypt hook" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Kernel parameters for encrypted boot
  KERNEL_PARAMS="cryptdevice=UUID=$ROOT_UUID:cryptroot root=/dev/mapper/cryptroot rootflags=subvol=@ rw"
else
  # For non-encrypted setups, use the root partition UUID directly
  ROOT_UUID=$(blkid -s UUID -o value "$ARCHUP_ROOT_PART")
  KERNEL_PARAMS="root=UUID=$ROOT_UUID rootflags=subvol=@ rw"
fi

# Add AMD-specific kernel parameters if configured
if [ -n "$ARCHUP_AMD_KERNEL_PARAMS" ]; then
  KERNEL_PARAMS="$KERNEL_PARAMS $ARCHUP_AMD_KERNEL_PARAMS"
  echo "Added AMD kernel params: $ARCHUP_AMD_KERNEL_PARAMS" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Install Limine to the disk (BIOS - optional, will fail on UEFI-only systems)
arch-chroot /mnt limine bios-install "$ARCHUP_DISK" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1 || true

# Create Limine configuration directory (Limine searches in /EFI/limine/)
mkdir -p /mnt/boot/EFI/BOOT

cat > /mnt/boot/limine.conf <<EOF
timeout: 0
default_entry: 0
interface_branding: ArchUp
interface_branding_colour: 6
graphics: yes
quiet: yes

:Arch Linux
    protocol: linux
    kernel_path: boot():/vmlinuz-linux
    module_path: boot():/initramfs-linux.img
    cmdline: $KERNEL_PARAMS

:Arch Linux (fallback)
    protocol: linux
    kernel_path: boot():/vmlinuz-linux
    module_path: boot():/initramfs-linux-fallback.img
    cmdline: $KERNEL_PARAMS
EOF

# Copy Limine EFI bootloader
cp /mnt/usr/share/limine/BOOTX64.EFI /mnt/boot/EFI/BOOT/

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Creating UEFI boot entry..."

# Detect EFI partition number (usually 1)
EFI_PART_NUM=$(echo "$ARCHUP_EFI_PART" | grep -o '[0-9]*$' | sed 's/^p//')

# Add UEFI boot entry using efibootmgr
# Limine doesn't automatically create NVRAM entry, we need to do it manually
arch-chroot /mnt efibootmgr --create \
  --disk "$ARCHUP_DISK" \
  --part "$EFI_PART_NUM" \
  --label "ArchUp" \
  --loader "\\EFI\\BOOT\\BOOTX64.EFI" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Limine installed and configured"

if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  echo "Installed Limine with encrypted root (UUID: $ROOT_UUID)" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  echo "Installed Limine with root UUID: $ROOT_UUID" >> "$ARCHUP_INSTALL_LOG_FILE"
fi
