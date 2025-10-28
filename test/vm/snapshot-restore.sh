#!/bin/bash
# Restore VM to a snapshot state
# Usage: ./snapshot-restore.sh [snapshot-name]
# Default snapshot name: clean-install
# WARNING: This will destroy current VM state!

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="$TEST_DIR/arch-test.qcow2"
SNAPSHOT_NAME="${1:-clean-install}"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

echo "Available snapshots:"
qemu-img snapshot -l "$IMAGE"
echo ""

read -p "Restore to snapshot '$SNAPSHOT_NAME'? This will DESTROY current VM state! (yes/NO): " -r
if [[ ! $REPLY =~ ^yes$ ]]; then
    echo "Aborted."
    exit 0
fi

echo "Restoring to snapshot: $SNAPSHOT_NAME"
qemu-img snapshot -a "$SNAPSHOT_NAME" "$IMAGE"

echo "[OK] VM successfully restored to '$SNAPSHOT_NAME'"
