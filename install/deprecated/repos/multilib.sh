#!/bin/bash
# Enable multilib in installed system

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring multilib repository..."

# Enable multilib in installed system's pacman.conf
enable_multilib "/mnt/etc/pacman.conf"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Multilib configured"
