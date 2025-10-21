#!/bin/bash
# Configure ISO pacman for faster downloads

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring pacman..."

# Enable ParallelDownloads on ISO
sed -i 's/^#ParallelDownloads = 5$/ParallelDownloads = 10/' /etc/pacman.conf

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Pacman configured for fast downloads"

echo "ISO Pacman configured: ParallelDownloads=10" >> "$ARCHUP_INSTALL_LOG_FILE"
