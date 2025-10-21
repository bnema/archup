#!/bin/bash
# ArchUp First Boot - Post-boot tasks that require running system

set -e

LOG_FILE="/var/log/archup-first-boot.log"

echo "=== ArchUp First Boot Setup - $(date) ===" > "$LOG_FILE"

# Run snapper configuration
if [ -f /usr/local/share/archup/post-boot/snapper.sh ]; then
  echo "Configuring snapper..." | tee -a "$LOG_FILE"
  bash /usr/local/share/archup/post-boot/snapper.sh >> "$LOG_FILE" 2>&1
fi

# Run SSH key generation
if [ -f /usr/local/share/archup/post-boot/ssh-keygen.sh ]; then
  echo "Generating SSH keys..." | tee -a "$LOG_FILE"
  bash /usr/local/share/archup/post-boot/ssh-keygen.sh >> "$LOG_FILE" 2>&1
fi

# Install archup-cli from GitHub
if [ -f /usr/local/share/archup/post-boot/archup-cli.sh ]; then
  echo "Installing archup-cli from GitHub..." | tee -a "$LOG_FILE"
  bash /usr/local/share/archup/post-boot/archup-cli.sh >> "$LOG_FILE" 2>&1
fi

echo "=== First Boot Setup Complete ===" >> "$LOG_FILE"

# Disable this service after first run
systemctl disable archup-first-boot.service

echo "First boot setup completed successfully"
