#!/bin/bash
# List all snapshots for the VM
# Usage: ./snapshot-list.sh

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="$TEST_DIR/arch-test.qcow2"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

echo "Snapshots for arch-test.qcow2:"
echo ""
qemu-img snapshot -l "$IMAGE"
