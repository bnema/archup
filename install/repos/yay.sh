#!/bin/bash
# Install yay AUR helper

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing yay..."
gum style --padding "0 0 1 $PADDING_LEFT" "This may take a few minutes..."

# Install base-devel if not already installed (required for building AUR packages)
arch-chroot /mnt pacman -S --noconfirm --needed base-devel git >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Get the username created during installation
USERNAME=$(arch-chroot /mnt ls /home | head -1)

if [[ -z "${USERNAME}" ]]; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Could not find user for yay installation"
  echo "ERROR: No user found for yay installation" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Build yay as user and install as root
echo "Building yay package..." >> "$ARCHUP_INSTALL_LOG_FILE"
if ! arch-chroot /mnt bash <<CHROOT_SCRIPT >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
set -e
set -u
set -o pipefail
cd /tmp
su ${USERNAME} -c "git clone https://aur.archlinux.org/yay-bin.git && cd yay-bin && makepkg -s --noconfirm"
cd yay-bin
pacman -U --noconfirm yay-bin-*.pkg.tar.zst
cd /tmp
rm -rf yay-bin
CHROOT_SCRIPT
then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] yay installation failed"
  echo "ERROR: yay installation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Verify installation
if arch-chroot /mnt su - "${USERNAME}" -c "yay --version" >/dev/null 2>&1; then
  gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] yay installed successfully"
  echo "AUR: enabled (yay installed for user: ${USERNAME})" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] yay verification failed"
  echo "ERROR: yay verification failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi
