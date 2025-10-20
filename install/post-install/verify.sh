#!/bin/bash
# Post-install verification: Ensure everything is configured correctly

#######################################
# Description: Verify installation components
# Globals:
#   ARCHUP_INSTALL_LOG_FILE
#   ARCHUP_ENCRYPTION
#   PADDING_LEFT
# Arguments:
#   None
# Outputs:
#   Verification status to log and terminal
# Returns:
#   0 if successful, 1 on critical failure
#######################################

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Verifying installation..."

# Track verification failures
VERIFICATION_FAILED=0

#######################################
# Description: Verify a file exists
# Arguments:
#   $1 - File path (without /mnt prefix)
#   $2 - Description
# Returns:
#   0 if exists, 1 if missing
#######################################
verify_file() {
  local file=$1
  local description=$2

  if [ -f "/mnt$file" ]; then
    echo "[OK] $description: $file" >> "$ARCHUP_INSTALL_LOG_FILE"
    return 0
  else
    echo "[ERROR] $description missing: $file" >> "$ARCHUP_INSTALL_LOG_FILE"
    VERIFICATION_FAILED=1
    return 1
  fi
}

verify_dir() {
  local dir=$1
  local description=$2

  if [ -d "/mnt$dir" ]; then
    echo "[OK] $description: $dir" >> "$ARCHUP_INSTALL_LOG_FILE"
    return 0
  else
    echo "[ERROR] $description missing: $dir" >> "$ARCHUP_INSTALL_LOG_FILE"
    VERIFICATION_FAILED=1
    return 1
  fi
}
verify_file "/boot/vmlinuz-linux" "Kernel"
verify_file "/boot/initramfs-linux.img" "Initramfs"
verify_file "/boot/initramfs-linux-fallback.img" "Fallback initramfs"

verify_file "/boot/EFI/limine/BOOTX64.EFI" "Limine bootloader"
verify_file "/boot/EFI/limine/limine.conf" "Limine configuration"

if [ -f "/mnt/boot/EFI/limine/limine.conf" ]; then
  if grep -q "^/" /mnt/boot/EFI/limine/limine.conf && \
     grep -q "protocol:" /mnt/boot/EFI/limine/limine.conf && \
     grep -q "path:" /mnt/boot/EFI/limine/limine.conf; then
    echo "[OK] Limine config validated" >> "$ARCHUP_INSTALL_LOG_FILE"
  else
    echo "[ERROR] Limine config validation failed" >> "$ARCHUP_INSTALL_LOG_FILE"
    VERIFICATION_FAILED=1
  fi
fi
verify_file "/etc/fstab" "Filesystem table"
verify_file "/etc/hostname" "Hostname"
verify_file "/etc/locale.conf" "Locale configuration"
verify_file "/etc/locale.gen" "Locale generation file"

if [ -f "/mnt/etc/systemd/network/20-wired.network" ] || [ -f "/mnt/etc/NetworkManager/NetworkManager.conf" ]; then
  echo "[OK] Network configuration present" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  echo "[ERROR] Network configuration missing" >> "$ARCHUP_INSTALL_LOG_FILE"
  VERIFICATION_FAILED=1
fi

if arch-chroot /mnt ls /home 2>/dev/null | grep -q .; then
  USERNAME=$(arch-chroot /mnt ls /home 2>/dev/null | head -1)
  echo "[OK] User created: $USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
  verify_dir "/home/$USERNAME" "User home directory"
else
  echo "[ERROR] No user account found" >> "$ARCHUP_INSTALL_LOG_FILE"
  VERIFICATION_FAILED=1
fi

if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
  if grep -q "encrypt" /mnt/etc/mkinitcpio.conf 2>/dev/null; then
    echo "[OK] Encryption hooks configured" >> "$ARCHUP_INSTALL_LOG_FILE"
  else
    echo "[ERROR] Encryption hooks missing in mkinitcpio.conf" >> "$ARCHUP_INSTALL_LOG_FILE"
    VERIFICATION_FAILED=1
  fi
fi

if [ -f "/mnt/boot/arch-logo.png" ]; then
  echo "[OK] Boot logo installed" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  echo "[WARN] Boot logo not installed (optional)" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Final verification result
echo "" >> "$ARCHUP_INSTALL_LOG_FILE"
if [ $VERIFICATION_FAILED -eq 0 ]; then
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] All critical components verified"
  echo "=== POST-INSTALL VERIFICATION: PASSED ===" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  gum style --foreground 1 --padding "0 0 1 $PADDING_LEFT" "[ERROR] Verification failed - check logs"
  echo "=== POST-INSTALL VERIFICATION: FAILED ===" >> "$ARCHUP_INSTALL_LOG_FILE"
  echo ""

  # Ask user if they want to continue despite errors
  if ! gum confirm "Continue with unmounting despite verification errors?" --default=false --padding "0 0 1 $PADDING_LEFT"; then
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "[ERROR] Installation aborted by user"
    echo "Installation aborted by user after verification failure" >> "$ARCHUP_INSTALL_LOG_FILE"
    exit 1
  fi

  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[WARN] Continuing despite verification errors"
  echo "User chose to continue despite verification errors" >> "$ARCHUP_INSTALL_LOG_FILE"
fi
