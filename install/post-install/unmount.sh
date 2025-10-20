#!/bin/bash
# Unmount all drives in preparation for reboot

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Unmounting partitions..."

# Unmount in reverse order (nested mounts first)
umount -R /mnt >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

if [ $? -eq 0 ]; then
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] All partitions unmounted"
  echo "Unmounted all partitions successfully" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Close encrypted volume if encryption was enabled
  if [ "$ARCHUP_ENCRYPTION" = "enabled" ]; then
    cryptsetup close cryptroot >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1 || true
    echo "Closed encrypted volume" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
else
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[WARN] Some partitions may still be mounted"
  echo "Warning: Failed to unmount all partitions cleanly" >> "$ARCHUP_INSTALL_LOG_FILE"
fi
