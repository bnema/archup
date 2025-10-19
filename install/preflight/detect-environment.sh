#!/bin/bash
# Detect keyboard layout and WiFi settings from the live ISO

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Detecting ISO environment..."

# Detect keyboard layout from current console
# Try to get the layout from loadkeys or localectl
DETECTED_KEYMAP=""

if command -v localectl >/dev/null 2>&1; then
  DETECTED_KEYMAP=$(localectl status | grep "VC Keymap" | awk '{print $3}')
fi

# Fallback: check /etc/vconsole.conf if it exists
if [ -z "$DETECTED_KEYMAP" ] && [ -f /etc/vconsole.conf ]; then
  DETECTED_KEYMAP=$(grep "^KEYMAP=" /etc/vconsole.conf | cut -d'=' -f2)
fi

# Default to us if nothing detected
if [ -z "$DETECTED_KEYMAP" ]; then
  DETECTED_KEYMAP="us"
fi

export ARCHUP_KEYMAP="$DETECTED_KEYMAP"
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ Keyboard layout: $ARCHUP_KEYMAP"
echo "Keyboard layout detected: $ARCHUP_KEYMAP" | tee -a "$ARCHUP_INSTALL_LOG_FILE"

# Detect WiFi connection from iwd
# Check if iwd is running and if we're connected to a network
WIFI_SSID=""
WIFI_PASSPHRASE=""

if command -v iwctl >/dev/null 2>&1; then
  # Get the first wireless device
  WIFI_DEVICE=$(iwctl device list | grep -oP 'wlan\d+' | head -1)

  if [ -n "$WIFI_DEVICE" ]; then
    # Check if connected to a network
    WIFI_SSID=$(iwctl station "$WIFI_DEVICE" show | grep "Connected network" | awk '{print $3}')

    if [ -n "$WIFI_SSID" ]; then
      # iwd stores credentials in /var/lib/iwd/
      # The PSK is stored in a file named SSID.psk
      IWD_CONFIG="/var/lib/iwd/${WIFI_SSID}.psk"

      if [ -f "$IWD_CONFIG" ]; then
        # Extract passphrase from iwd config (stored in [Security] section as PreSharedKey)
        WIFI_PASSPHRASE=$(grep "^PreSharedKey=" "$IWD_CONFIG" | cut -d'=' -f2)
      fi

      export ARCHUP_WIFI_SSID="$WIFI_SSID"
      export ARCHUP_WIFI_DEVICE="$WIFI_DEVICE"
      export ARCHUP_WIFI_PASSPHRASE="$WIFI_PASSPHRASE"

      gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ WiFi network: $WIFI_SSID (device: $WIFI_DEVICE)"
      echo "WiFi detected: SSID=$WIFI_SSID, device=$WIFI_DEVICE" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
    else
      gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "⊘ No active WiFi connection detected"
      echo "WiFi: not connected" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
    fi
  else
    gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "⊘ No WiFi device found"
    echo "WiFi: no device" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
  fi
else
  gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "⊘ iwd not available"
  echo "WiFi: iwd not available" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
fi

echo
