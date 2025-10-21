#!/bin/bash
# Install base system with pacstrap

# Read base package list
mapfile -t packages < <(grep -v '^#' "$ARCHUP_INSTALL/base.packages" | grep -v '^$')

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

# Install base system with spinner
echo "Installing ${#packages[@]} packages: ${packages[*]}" >> "$ARCHUP_INSTALL_LOG_FILE"
if ! run_with_spinner "Installing base system (${#packages[@]} packages)..." "pacstrap /mnt ${packages[*]}"; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Base system installation failed"
  exit 1
fi

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
