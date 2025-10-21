#!/bin/bash
# AUR helper installation - consolidated paru and yay

# Function to install paru (available in chaotic-aur)
install_paru() {
  local username="$1"
  echo "Installing paru from chaotic-aur..." >> "$ARCHUP_INSTALL_LOG_FILE"

  if ! run_with_spinner "Installing paru AUR helper..." "arch-chroot /mnt pacman -S --noconfirm paru"; then
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] paru installation failed"
    echo "ERROR: paru installation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
    export ARCHUP_AUR="failed"
    return 1
  fi

  # Verify installation
  if arch-chroot /mnt su - "${username}" -c "paru --version" >/dev/null 2>&1; then
    gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] paru installed successfully"
    echo "AUR: enabled (paru installed for user: ${username})" >> "$ARCHUP_INSTALL_LOG_FILE"
  else
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] paru verification failed"
    echo "ERROR: paru verification failed" >> "$ARCHUP_INSTALL_LOG_FILE"
    export ARCHUP_AUR="failed"
    return 1
  fi
}

# Function to install yay (available in chaotic-aur)
install_yay() {
  local username="$1"
  echo "Installing yay from chaotic-aur..." >> "$ARCHUP_INSTALL_LOG_FILE"

  if ! run_with_spinner "Installing yay AUR helper..." "arch-chroot /mnt pacman -S --noconfirm yay"; then
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] yay installation failed"
    echo "ERROR: yay installation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
    export ARCHUP_AUR="failed"
    return 1
  fi

  # Verify installation
  if arch-chroot /mnt su - "${username}" -c "yay --version" >/dev/null 2>&1; then
    gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] yay installed successfully"
    echo "AUR: enabled (yay installed for user: ${username})" >> "$ARCHUP_INSTALL_LOG_FILE"
  else
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] yay verification failed"
    echo "ERROR: yay verification failed" >> "$ARCHUP_INSTALL_LOG_FILE"
    export ARCHUP_AUR="failed"
    return 1
  fi
}

# Main execution
# AUR Support prompt
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AUR Support"
echo

if ! gum confirm "Enable AUR support?" --padding "0 0 1 $PADDING_LEFT"; then
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[SKIP] Skipping AUR support"
  echo "AUR: disabled" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="disabled"
  return 0
fi

export ARCHUP_AUR="enabled"

# Choose AUR helper
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Choose AUR Helper"
AUR_HELPER=$(gum choose --cursor.foreground 6 --padding "0 0 1 $PADDING_LEFT" "paru" "yay")

# Get the username created during installation
USERNAME=$(arch-chroot /mnt ls /home | head -1)

if [[ -z "${USERNAME}" ]]; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Could not find user for AUR helper installation"
  echo "ERROR: No user found for AUR helper installation" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="failed"
  return 1
fi

# Install selected AUR helper (both available in chaotic-aur)
if [ "$AUR_HELPER" = "paru" ]; then
  install_paru "$USERNAME"
else
  install_yay "$USERNAME"
fi
