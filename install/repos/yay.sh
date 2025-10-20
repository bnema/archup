#!/bin/bash
# Install yay AUR helper

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AUR Support"
echo

# Ask if user wants AUR support
if ! gum confirm "Enable AUR support (install yay)?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[SKIP] Skipping AUR support"
  echo "AUR: disabled" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="disabled"
  return 0
fi

export ARCHUP_AUR="enabled"

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing yay AUR helper..."
gum style --padding "0 0 1 $PADDING_LEFT" "This may take a few minutes..."

# Install base-devel if not already installed (required for building AUR packages)
arch-chroot /mnt pacman -S --noconfirm --needed base-devel git >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Get the username created during installation
USERNAME=$(arch-chroot /mnt ls /home | head -1)

if [ -z "$USERNAME" ]; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Error: Could not find user for yay installation"
  echo "ERROR: No user found for yay installation" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Clone yay-bin from AUR as the user (not root)
arch-chroot /mnt su - "$USERNAME" -c "cd /tmp && git clone https://aur.archlinux.org/yay-bin.git" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1 || true

# Build and install yay as the user
arch-chroot /mnt su - "$USERNAME" -c "cd /tmp/yay-bin && makepkg -si --noconfirm" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1 || true

# Clean up
arch-chroot /mnt rm -rf /tmp/yay-bin >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Verify installation
if arch-chroot /mnt su - "$USERNAME" -c "yay --version" >/dev/null 2>&1; then
  gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] yay installed successfully"
  echo "AUR: enabled (yay installed for user: $USERNAME)" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] yay installation failed"
  echo "ERROR: yay installation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi
