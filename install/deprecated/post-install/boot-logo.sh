#!/bin/bash
# Download and configure Limine boot logo

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing boot logo..."

# Download logo directly to boot partition (no chroot needed, boot is mounted at /mnt/boot)
curl -fsSL "$ARCHUP_RAW_URL/assets/Arch_Linux__Crystal__icon.png" \
  -o /mnt/boot/arch-logo.png >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

if [ -f /mnt/boot/arch-logo.png ]; then
  if [ -f /mnt/boot/EFI/limine/limine.conf ]; then
    sed -i '/^graphics: yes$/a wallpaper: boot():/arch-logo.png\nwallpaper_style: centered\nbackdrop: 000000' \
      /mnt/boot/EFI/limine/limine.conf
    gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Boot logo installed"
  fi
fi
