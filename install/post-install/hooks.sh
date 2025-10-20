#!/bin/bash
# Install pacman hooks for automatic system maintenance

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Installing pacman hooks..."

mkdir -p /mnt/etc/pacman.d/hooks

# Limine bootloader auto-update hook
cat > /mnt/etc/pacman.d/hooks/99-limine.hook <<'EOF'
[Trigger]
Operation = Install
Operation = Upgrade
Type = Package
Target = limine

[Action]
Description = Deploying Limine after upgrade...
When = PostTransaction
Exec = /usr/bin/cp /usr/share/limine/BOOTX64.EFI /boot/EFI/limine/
EOF

echo "Installed Limine update hook" >> "$ARCHUP_INSTALL_LOG_FILE"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Pacman hooks installed"
