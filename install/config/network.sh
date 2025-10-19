#!/bin/bash
# Configure network (systemd-networkd + systemd-resolved)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring network..."

# Enable systemd-networkd for DHCP
mkdir -p /mnt/etc/systemd/network

cat > /mnt/etc/systemd/network/20-wired.network <<EOF
[Match]
Name=en*

[Network]
DHCP=yes
EOF

cat > /mnt/etc/systemd/network/25-wireless.network <<EOF
[Match]
Name=wl*

[Network]
DHCP=yes
EOF

# Enable services
arch-chroot /mnt systemctl enable systemd-networkd
arch-chroot /mnt systemctl enable systemd-resolved

# Enable iwd for WiFi
arch-chroot /mnt systemctl enable iwd

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Network configured (DHCP)"

echo "Enabled systemd-networkd and iwd" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
