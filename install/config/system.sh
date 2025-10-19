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
arch-chroot /mnt ln -sf "/usr/share/zoneinfo/$ARCHUP_TIMEZONE" /etc/localtime
arch-chroot /mnt hwclock --systohc

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "✓ Timezone set to: $ARCHUP_TIMEZONE"

# Set locale to en_US.UTF-8
echo "en_US.UTF-8 UTF-8" > /mnt/etc/locale.gen
arch-chroot /mnt locale-gen
echo "LANG=en_US.UTF-8" > /mnt/etc/locale.conf

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Locale set to: en_US.UTF-8"

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

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "✓ Hostname set to: $ARCHUP_HOSTNAME"

echo "Timezone: $ARCHUP_TIMEZONE" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Locale: en_US.UTF-8" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
echo "Hostname: $ARCHUP_HOSTNAME" | tee -a "$ARCHUP_INSTALL_LOG_FILE"
