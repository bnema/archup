#!/bin/bash
# Configure Plymouth splash screen with ArchUp theme

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring Plymouth splash screen..."

THEME_NAME="archup"
THEME_DIR="/mnt/usr/share/plymouth/themes/$THEME_NAME"

# Create theme directory
mkdir -p "$THEME_DIR"

# Download all Plymouth theme files from GitHub
PLYMOUTH_FILES=(
  "archup.plymouth"
  "archup.script"
  "logo.png"
  "bullet.png"
  "entry.png"
  "lock.png"
  "progress_bar.png"
  "progress_box.png"
)

for file in "${PLYMOUTH_FILES[@]}"; do
  curl -fsSL "$ARCHUP_RAW_URL/assets/plymouth/$file" \
    -o "$THEME_DIR/$file" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
done

# Set permissions
chmod 644 "$THEME_DIR"/*

# Set as default theme
arch-chroot /mnt plymouth-set-default-theme "$THEME_NAME" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Regenerate initramfs with Plymouth
arch-chroot /mnt mkinitcpio -P >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Plymouth configured"
