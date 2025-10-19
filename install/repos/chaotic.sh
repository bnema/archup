#!/bin/bash
# Add Chaotic-AUR repository

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Chaotic-AUR Repository"
echo

# Ask if user wants Chaotic-AUR
if ! gum confirm "Enable Chaotic-AUR repository?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[SKIP] Skipping Chaotic-AUR"
  echo "Chaotic-AUR: disabled"
  export ARCHUP_CHAOTIC="disabled"
  return 0
fi

export ARCHUP_CHAOTIC="enabled"

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Adding Chaotic-AUR repository..."
gum style --padding "0 0 1 $PADDING_LEFT" "Installing keyring and mirrorlist..."

# Install Chaotic-AUR keyring and mirrorlist
arch-chroot /mnt pacman-key --recv-key 3056513887B78AEB --keyserver keyserver.ubuntu.com
arch-chroot /mnt pacman-key --lsign-key 3056513887B78AEB

# Install chaotic-keyring and chaotic-mirrorlist
arch-chroot /mnt pacman -U --noconfirm \
  'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst' \
  'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst'

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
arch-chroot /mnt pacman -Sy

# Verify Chaotic-AUR is working
if arch-chroot /mnt pacman -Sl chaotic-aur >/dev/null 2>&1; then
  gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Chaotic-AUR enabled successfully"
  echo "Chaotic-AUR: enabled (repo added to pacman.conf)"
else
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Chaotic-AUR verification failed"
  echo "WARNING: Chaotic-AUR verification failed"
  export ARCHUP_CHAOTIC="failed"
  return 1
fi
