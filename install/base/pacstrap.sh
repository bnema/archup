#!/bin/bash
# Install base system with pacstrap

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing base system..."
gum style --padding "0 0 1 $PADDING_LEFT" "This may take a few minutes..."

# Read barebone package list
mapfile -t packages < <(grep -v '^#' "$ARCHUP_INSTALL/presets/barebone.packages" | grep -v '^$')

# Add cryptsetup if encryption is enabled
if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  packages+=("cryptsetup")
  echo "Adding cryptsetup for LUKS encryption" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
fi

# Install base system
pacstrap /mnt "${packages[@]}"

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "✓ Base system installed"

echo "Installed ${#packages[@]} packages" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
