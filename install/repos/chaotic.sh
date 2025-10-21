#!/bin/bash
# Add Chaotic-AUR repository (required for limine-snapper-sync and other packages)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Adding Chaotic-AUR repository..."

export ARCHUP_CHAOTIC="enabled"

# Install Chaotic-AUR keyring and mirrorlist
arch-chroot /mnt pacman-key --recv-key 3056513887B78AEB --keyserver keyserver.ubuntu.com >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt pacman-key --lsign-key 3056513887B78AEB >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Install chaotic-keyring and chaotic-mirrorlist
arch-chroot /mnt pacman -U --noconfirm \
  'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst' \
  'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst' >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Add Chaotic-AUR to pacman.conf
if ! grep -q "\[chaotic-aur\]" /mnt/etc/pacman.conf; then
  cat >> /mnt/etc/pacman.conf <<EOF

# Chaotic-AUR repository
[chaotic-aur]
Include = /etc/pacman.d/chaotic-mirrorlist
EOF

  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Added Chaotic-AUR to pacman.conf"
fi

# Update package databases
arch-chroot /mnt pacman -Sy >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Verify Chaotic-AUR is working
if arch-chroot /mnt pacman -Sl chaotic-aur >/dev/null 2>&1; then
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Chaotic-AUR enabled successfully"
  echo "Chaotic-AUR: enabled (repo added to pacman.conf)" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Chaotic-AUR verification failed"
  echo "WARNING: Chaotic-AUR verification failed" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_CHAOTIC="failed"
  return 1
fi

# Install extra packages from Chaotic-AUR
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing extra packages..."

# Read extra package list
mapfile -t extra_packages < <(grep -v '^#' "$ARCHUP_INSTALL/extra.packages" | grep -v '^$')

echo "Installing ${#extra_packages[@]} extra packages: ${extra_packages[*]}" >> "$ARCHUP_INSTALL_LOG_FILE"

if ! run_with_spinner "Installing extra packages (${#extra_packages[@]} packages)..." "arch-chroot /mnt pacman -S --noconfirm ${extra_packages[*]}"; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Extra packages installation failed"
  echo "WARNING: Some extra packages failed to install" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Extra packages installed"
  echo "Installed ${#extra_packages[@]} extra packages" >> "$ARCHUP_INSTALL_LOG_FILE"
fi
