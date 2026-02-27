#!/bin/bash
# dms-opt-in.sh — Optional Dank Linux full suite installation prompt
# Runs at end of first-boot setup. Asks user if they want the full
# Dank Linux desktop (compositor + DMS shell + terminal + all tools).
set -euo pipefail

LOG_FILE="/var/log/archup-first-boot.log"

log() { echo "[dms-opt-in] $*" | tee -a "$LOG_FILE"; }

FLAG_FILE="/var/lib/archup-install-danklinux"

# Auto-install if flag was set during installation
if [ -f "$FLAG_FILE" ]; then
  log "Dank Linux install flag detected — running installer automatically."
  echo ""
  echo "Installing Dank Linux (selected during system installation)..."
  if curl -fsSL https://install.danklinux.com | sh; then
    rm -f "$FLAG_FILE"
    log "Dank Linux installation complete."
  else
    log "Dank Linux installer failed — flag preserved at $FLAG_FILE for retry."
    exit 1
  fi
  exit 0
fi

# Skip if not running in an interactive terminal
if [ ! -t 0 ]; then
  log "Not running in interactive terminal — skipping Dank Linux opt-in."
  exit 0
fi

# Manual prompt fallback (for reinstall / re-run scenarios)
echo ""
echo "=============================================="
echo "       Dank Linux — Optional Install"
echo "=============================================="
echo ""
echo "  Dank Linux is a full Wayland desktop suite:"
echo "    - Compositor: niri or Hyprland (your choice)"
echo "    - Shell: DankMaterialShell (Quickshell/Go)"
echo "    - Terminal: Ghostty"
echo "    - Auto-theming, notifications, app launcher,"
echo "      lock screen, system tray, search — all"
echo "      integrated. One install. Everything works."
echo ""
echo "  Install: curl -fsSL https://install.danklinux.com | sh"
echo "  Note: This runs the official Dank Linux installer. Review at:"
echo "  https://github.com/AvengeMedia/DankMaterialShell"
echo ""

read -r -p "  Do you want to run the Dank Linux installer? [y/N] " REPLY || true
echo ""

case "$REPLY" in
  [yY][eE][sS]|[yY])
    log "User opted in — starting Dank Linux installer."
    echo "Starting Dank Linux installer..."
    curl -fsSL https://install.danklinux.com | sh
    log "Dank Linux installation complete."
    ;;
  *)
    log "User skipped Dank Linux installation."
    echo "Skipping. Install later with:"
    echo "  curl -fsSL https://install.danklinux.com | sh"
    echo ""
    ;;
esac
