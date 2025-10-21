#!/bin/bash
# Helper function to enable multilib repository

#######################################
# Description: Enable multilib in pacman.conf
# Arguments:
#   $1 - Path to pacman.conf (optional, defaults to /etc/pacman.conf)
# Returns:
#   0 if successful, 1 on error
#######################################
enable_multilib() {
  local pacman_conf="${1:-/etc/pacman.conf}"

  # Check if multilib is already enabled
  if grep -q "^\[multilib\]" "$pacman_conf"; then
    echo "Multilib already enabled in $pacman_conf" >> "$ARCHUP_INSTALL_LOG_FILE"
    return 0
  fi

  # Uncomment multilib section
  sed -i '/^#\[multilib\]/,/^#Include = \/etc\/pacman.d\/mirrorlist/ s/^#//' "$pacman_conf"

  echo "Multilib enabled in $pacman_conf" >> "$ARCHUP_INSTALL_LOG_FILE"
  return 0
}
