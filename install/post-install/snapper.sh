#!/bin/bash
# Configure Snapper and Limine integration for btrfs snapshots

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring btrfs snapshots..."

arch-chroot /mnt pacman -S --noconfirm --needed limine-snapper-sync >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

mkdir -p /mnt/usr/local/share/archup/post-boot

cp "$ARCHUP_INSTALL/post-boot/all.sh" /mnt/usr/local/share/archup/post-boot/
cp "$ARCHUP_INSTALL/post-boot/snapper.sh" /mnt/usr/local/share/archup/post-boot/
cp "$ARCHUP_INSTALL/post-boot/ssh-keygen.sh" /mnt/usr/local/share/archup/post-boot/
chmod +x /mnt/usr/local/share/archup/post-boot/*.sh

sed "s|__ARCHUP_EMAIL__|$(config_get "ARCHUP_EMAIL")|g; s|__ARCHUP_USERNAME__|$(config_get "ARCHUP_USERNAME")|g" \
  "$ARCHUP_INSTALL/post-boot/archup-first-boot.service" > /mnt/etc/systemd/system/archup-first-boot.service

arch-chroot /mnt systemctl enable archup-first-boot.service >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

LIMINE_CONFIG="/mnt/boot/EFI/limine/limine.conf"
CMDLINE=$(grep "^[[:space:]]*cmdline:" "$LIMINE_CONFIG" | head -1 | sed 's/^[[:space:]]*cmdline:[[:space:]]*//')
cat > /mnt/etc/default/limine <<EOF
TARGET_OS_NAME="ArchUp"

ESP_PATH="/boot"

KERNEL_CMDLINE[default]="$CMDLINE"

ENABLE_UKI=no
ENABLE_LIMINE_FALLBACK=yes

FIND_BOOTLOADERS=yes

BOOT_ORDER="*, *fallback, Snapshots"

MAX_SNAPSHOT_ENTRIES=5

SNAPSHOT_FORMAT_CHOICE=5
EOF

arch-chroot /mnt systemctl enable limine-snapper-sync.service >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Btrfs snapshots configured"
