#!/bin/bash
# Configure UFW (Uncomplicated Firewall) with basic protection

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring firewall..."

# UFW is installed via extra.packages
# Configure basic firewall rules

# Set default policies
arch-chroot /mnt ufw --force default deny incoming >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt ufw --force default allow outgoing >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Allow common local services (optional - can be removed for stricter security)
# arch-chroot /mnt ufw allow ssh >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Enable UFW
arch-chroot /mnt ufw --force enable >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Enable UFW service to start on boot
arch-chroot /mnt systemctl enable ufw.service >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Firewall configured (default deny incoming)"

echo "UFW configured: default deny incoming, allow outgoing" >> "$ARCHUP_INSTALL_LOG_FILE"
