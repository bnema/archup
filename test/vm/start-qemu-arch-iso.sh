#!/bin/bash
# Start Arch ISO in QEMU with UEFI support and SSH access

set -e

# Get test directory
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ISO_FILE="$TEST_DIR/archlinux-x86_64.iso"
ISO_URL="https://geo.mirror.pkgbuild.com/iso/latest/archlinux-x86_64.iso"

# Download latest Arch ISO if not present
if [[ ! -f "$ISO_FILE" ]]; then
    echo "Arch Linux ISO not found in $TEST_DIR"
    echo "Downloading latest Arch ISO from $ISO_URL..."
    curl -L -o "$ISO_FILE" "$ISO_URL"
    echo "Download complete!"
else
    echo "Using existing ISO: $ISO_FILE"
fi

qemu-system-x86_64 \
  -enable-kvm \
  -m 2048 \
  -smp 2 \
  -bios /usr/share/edk2/x64/OVMF.4m.fd \
  -boot d \
  -cdrom "$ISO_FILE" \
  -drive file="$TEST_DIR/arch-test.qcow2",format=qcow2,if=virtio \
  -display gtk \
  -vga virtio \
  -net nic,model=virtio \
  -net user,hostfwd=tcp::2222-:22
