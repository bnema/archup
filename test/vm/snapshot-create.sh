#!/bin/bash
# Create a snapshot of the installed VM
# Usage: ./snapshot-create.sh [snapshot-name]
# Default snapshot name: clean-install

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="$TEST_DIR/arch-test.qcow2"
SNAPSHOT_NAME="${1:-clean-install}"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

echo "Creating snapshot: $SNAPSHOT_NAME"
qemu-img snapshot -c "$SNAPSHOT_NAME" "$IMAGE"

echo "✓ Snapshot '$SNAPSHOT_NAME' created successfully"
echo ""
echo "Available snapshots:"
qemu-img snapshot -l "$IMAGE"
