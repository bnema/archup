#!/bin/bash
# Start installed Arch system in QEMU with UEFI support and SSH access

# Get test directory
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

qemu-system-x86_64 \
  -enable-kvm \
  -m 2048 \
  -smp 2 \
  -bios /usr/share/edk2/x64/OVMF.4m.fd \
  -drive file="$TEST_DIR/arch-test.qcow2",format=qcow2,if=virtio \
  -display gtk \
  -vga virtio \
  -net nic,model=virtio \
  -net user,hostfwd=tcp::2222-:22
