#!/bin/bash
# Start installed Arch system in QEMU with realistic hardware profile
# Usage: ./start-qemu-installed.sh [--profile desktop|laptop] [--secure-boot on|off] [--ssh-port PORT]
set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

source "$TEST_DIR/lib/qemu-profile.sh"
build_qemu_args "$PROFILE" "$SECURE_BOOT" "$SSH_PORT" "$TEST_DIR"

exec qemu-system-x86_64 "${QEMU_ARGS[@]}"
