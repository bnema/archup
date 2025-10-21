#!/bin/bash
# Configure pacman for better UX

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring pacman..."

# Enable Color output
sed -i 's/^#Color$/Color/' /mnt/etc/pacman.conf

# Enable ParallelDownloads
sed -i 's/^#ParallelDownloads = 5$/ParallelDownloads = 5/' /mnt/etc/pacman.conf

# Enable ILoveCandy (Pac-Man progress bar)
sed -i '/^Color$/a ILoveCandy' /mnt/etc/pacman.conf

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Pacman configured (color, parallel downloads, candy)"

echo "Pacman configured: Color, ParallelDownloads=5, ILoveCandy" >> "$ARCHUP_INSTALL_LOG_FILE"
