#!/bin/bash
# Configure ISO pacman for faster downloads

# Enable ParallelDownloads on ISO
sed -i 's/^#ParallelDownloads = 5$/ParallelDownloads = 10/' /etc/pacman.conf

echo "ISO Pacman configured: ParallelDownloads=10" >> "$ARCHUP_INSTALL_LOG_FILE"
