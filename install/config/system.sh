#!/bin/bash
# System configuration (timezone, locale, hostname)

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "System Configuration"
echo

# Ask for timezone
ARCHUP_TIMEZONE=$(gum input --placeholder "Enter timezone (e.g., America/New_York)" \
  --prompt "Timezone: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ -z "$ARCHUP_TIMEZONE" ]; then
  ARCHUP_TIMEZONE="UTC"
fi

# Set timezone
arch-chroot /mnt ln -sf "/usr/share/zoneinfo/$ARCHUP_TIMEZONE" /etc/localtime >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt hwclock --systohc >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Timezone set to: $ARCHUP_TIMEZONE"

# Set locale to en_US.UTF-8
echo "en_US.UTF-8 UTF-8" > /mnt/etc/locale.gen
arch-chroot /mnt locale-gen >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
echo "LANG=en_US.UTF-8" > /mnt/etc/locale.conf

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Locale set to: en_US.UTF-8"

# Set console keyboard layout (detected from ISO)
if [ -n "$ARCHUP_KEYMAP" ]; then
  echo "KEYMAP=$ARCHUP_KEYMAP" > /mnt/etc/vconsole.conf
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Console keymap set to: $ARCHUP_KEYMAP"
  echo "Console keymap: $ARCHUP_KEYMAP" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Ask for hostname
ARCHUP_HOSTNAME=$(gum input --placeholder "Enter hostname" \
  --prompt "Hostname: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ -z "$ARCHUP_HOSTNAME" ]; then
  ARCHUP_HOSTNAME="archlinux"
fi

# Set hostname
echo "$ARCHUP_HOSTNAME" > /mnt/etc/hostname

# Configure hosts file
cat > /mnt/etc/hosts <<EOF
127.0.0.1   localhost
::1         localhost
127.0.1.1   $ARCHUP_HOSTNAME.localdomain $ARCHUP_HOSTNAME
EOF

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Hostname set to: $ARCHUP_HOSTNAME"

echo "Timezone: $ARCHUP_TIMEZONE" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Locale: en_US.UTF-8" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Hostname: $ARCHUP_HOSTNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
