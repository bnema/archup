#!/bin/bash
# Bootloader phase orchestrator
# Phase 1: systemd-boot only (Limine support comes in Phase 3)

if [ "$ARCHUP_BOOTLOADER" = "systemd-boot" ]; then
  run_logged "$ARCHUP_INSTALL/boot/systemd-boot.sh"
else
  # For Phase 1, we only support systemd-boot
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Note: Limine support coming in Phase 3. Installing systemd-boot..."
  run_logged "$ARCHUP_INSTALL/boot/systemd-boot.sh"
fi
