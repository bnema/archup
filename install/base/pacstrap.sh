#!/bin/bash
# Install base system with pacstrap

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing base system..."
gum style --padding "0 0 1 $PADDING_LEFT" "This may take a few minutes..."

# Read barebone package list
mapfile -t packages < <(grep -v '^#' "$ARCHUP_INSTALL/presets/barebone.packages" | grep -v '^$')

# Add kernel (selected in kernel.sh)
if [ -n "$ARCHUP_KERNEL" ]; then
  packages+=("$ARCHUP_KERNEL")
  echo "Adding kernel: $ARCHUP_KERNEL" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Add microcode (detected in kernel.sh)
if [ -n "$ARCHUP_MICROCODE" ]; then
  packages+=("$ARCHUP_MICROCODE")
  echo "Adding microcode: $ARCHUP_MICROCODE" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Add cryptsetup if encryption is enabled
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  packages+=("cryptsetup")
  echo "Adding cryptsetup for LUKS encryption" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Install base system
pacstrap /mnt "${packages[@]}"

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Base system installed"

echo "Installed ${#packages[@]} packages" >> "$ARCHUP_INSTALL_LOG_FILE"

# Configure CachyOS repository in installed system if linux-cachyos was selected
if [[ "$ARCHUP_KERNEL" == "linux-cachyos" ]]; then
  echo "Configuring CachyOS repository in installed system..." >> "$ARCHUP_INSTALL_LOG_FILE"

  # Add CachyOS repo before [core] repository
  if ! grep -q "\[cachyos\]" /mnt/etc/pacman.conf; then
    sed -i '/^\[core\]/i \
# CachyOS repositories\
[cachyos]\
Include = /etc/pacman.d/cachyos-mirrorlist\
' /mnt/etc/pacman.conf
  fi

  # Copy mirrorlist to installed system
  mkdir -p /mnt/etc/pacman.d
  cp /etc/pacman.d/cachyos-mirrorlist /mnt/etc/pacman.d/cachyos-mirrorlist

  echo "CachyOS repository configured in installed system" >> "$ARCHUP_INSTALL_LOG_FILE"
fi
