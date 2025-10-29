#!/bin/bash
# Delete a snapshot
# Usage: ./snapshot-delete.sh <snapshot-name>

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="$TEST_DIR/arch-test.qcow2"
SNAPSHOT_NAME="${1}"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

if [ -z "$SNAPSHOT_NAME" ]; then
    echo "Usage: $0 <snapshot-name>"
    echo ""
    echo "Available snapshots:"
    qemu-img snapshot -l "$IMAGE"
    exit 1
fi

read -p "Delete snapshot '$SNAPSHOT_NAME'? (yes/NO): " -r
if [[ ! $REPLY =~ ^yes$ ]]; then
    echo "Aborted."
    exit 0
fi

echo "Deleting snapshot: $SNAPSHOT_NAME"
qemu-img snapshot -d "$SNAPSHOT_NAME" "$IMAGE"

echo "✓ Snapshot '$SNAPSHOT_NAME' deleted successfully"
