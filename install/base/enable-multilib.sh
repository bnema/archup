#!/bin/bash
# Enable multilib on ISO before pacstrap

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Enabling multilib repository..."

# Enable multilib in ISO's pacman.conf
enable_multilib "/etc/pacman.conf"

# Update package database
if ! run_with_spinner "Updating package database..." "pacman -Sy"; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Failed to update package database"
  exit 1
fi

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Multilib enabled"
