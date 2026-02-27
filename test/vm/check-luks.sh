#!/bin/bash
# Check LUKS setup on the VM disk
# Usage: ./check-luks.sh [--disk DEVICE]  (default: /dev/nvme0n1p2)

DISK="/dev/nvme0n1p2"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --disk) DISK="$2"; shift 2 ;;
    *) echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
done

echo "Checking LUKS volume on $DISK..."
cryptsetup luksDump "$DISK" || { echo "luksDump failed — is $DISK a LUKS volume?" >&2; exit 1; }

echo ""
echo "Trying to open with 'test' password..."
echo "test" | cryptsetup open --test-passphrase "$DISK" && echo "Password 'test' works!" || echo "Password 'test' does NOT work!"
