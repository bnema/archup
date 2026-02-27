#!/bin/bash
# Start Arch ISO in QEMU with realistic hardware profile
# Usage: ./start-qemu-arch-iso.sh [--profile desktop|laptop] [--secure-boot on|off] [--ssh-port PORT]
set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ISO_FILE="$TEST_DIR/archlinux-x86_64.iso"
ISO_URL="https://geo.mirror.pkgbuild.com/iso/latest/archlinux-x86_64.iso"

PROFILE="desktop"
SECURE_BOOT="off"
SSH_PORT="2222"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile) PROFILE="$2"; shift 2 ;;
    --secure-boot) SECURE_BOOT="$2"; shift 2 ;;
    --ssh-port) SSH_PORT="$2"; shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

if [[ ! -f "$ISO_FILE" ]]; then
  echo "Arch Linux ISO not found in $TEST_DIR"
  echo "Downloading latest Arch ISO from $ISO_URL..."
  curl -L -o "$ISO_FILE" "$ISO_URL"
  echo "Download complete!"
else
  echo "Using existing ISO: $ISO_FILE"
fi

source "$TEST_DIR/lib/qemu-profile.sh"
build_qemu_args "$PROFILE" "$SECURE_BOOT" "$SSH_PORT" "$TEST_DIR"

# ISO boot additions
QEMU_ARGS+=(-boot d)
QEMU_ARGS+=(-cdrom "$ISO_FILE")

exec qemu-system-x86_64 "${QEMU_ARGS[@]}"
