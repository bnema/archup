#!/bin/bash
# Install paru AUR helper

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing paru..."
gum style --padding "0 0 1 $PADDING_LEFT" "This may take a few minutes..."

# Install base-devel if not already installed (required for building AUR packages)
arch-chroot /mnt pacman -S --noconfirm --needed base-devel git >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Get the username created during installation
USERNAME=$(arch-chroot /mnt ls /home | head -1)

if [[ -z "${USERNAME}" ]]; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Could not find user for paru installation"
  echo "ERROR: No user found for paru installation" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Build paru as user and install as root
echo "Building paru package..." >> "$ARCHUP_INSTALL_LOG_FILE"
if ! arch-chroot /mnt bash <<CHROOT_SCRIPT >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
set -e
set -u
set -o pipefail
cd /tmp
su ${USERNAME} -c "git clone https://aur.archlinux.org/paru.git && cd paru && makepkg -s --noconfirm"
cd paru
pacman -U --noconfirm paru-*.pkg.tar.zst
cd /tmp
rm -rf paru
CHROOT_SCRIPT
then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] paru installation failed"
  echo "ERROR: paru installation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Verify installation
if arch-chroot /mnt su - "${USERNAME}" -c "paru --version" >/dev/null 2>&1; then
  gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] paru installed successfully"
  echo "AUR: enabled (paru installed for user: ${USERNAME})" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] paru verification failed"
  echo "ERROR: paru verification failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi
