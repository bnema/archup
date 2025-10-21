#!/bin/bash
# Configure Snapper and Limine integration for btrfs snapshots

# Install limine-snapper-sync package
if ! run_with_spinner "Installing limine-snapper-sync..." "arch-chroot /mnt pacman -S --noconfirm --needed limine-snapper-sync"; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Failed to install limine-snapper-sync"
  return 1
fi

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
