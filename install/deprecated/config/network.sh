#!/bin/bash
# Configure network (NetworkManager)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring network..."

# Enable NetworkManager service
arch-chroot /mnt systemctl enable NetworkManager >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Migrate WiFi credentials from ISO if detected
if [ -n "$ARCHUP_WIFI_SSID" ] && [ -n "$ARCHUP_WIFI_PASSPHRASE" ]; then
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Migrating WiFi credentials..."

  # Create NetworkManager connection file
  cat > "/mnt/etc/NetworkManager/system-connections/${ARCHUP_WIFI_SSID}.nmconnection" <<EOF
[connection]
id=${ARCHUP_WIFI_SSID}
uuid=$(uuidgen)
type=wifi
autoconnect=true

[wifi]
mode=infrastructure
ssid=${ARCHUP_WIFI_SSID}

[wifi-security]
key-mgmt=wpa-psk
psk=${ARCHUP_WIFI_PASSPHRASE}

[ipv4]
method=auto

[ipv6]
addr-gen-mode=default
method=auto

[proxy]
EOF

  # Set correct permissions (NetworkManager requires 600)
  chmod 600 "/mnt/etc/NetworkManager/system-connections/${ARCHUP_WIFI_SSID}.nmconnection"

  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] WiFi credentials migrated: $ARCHUP_WIFI_SSID"
  echo "WiFi credentials migrated: $ARCHUP_WIFI_SSID (auto-connect enabled)" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Network configured"

echo "Enabled NetworkManager" >> "$ARCHUP_INSTALL_LOG_FILE"
