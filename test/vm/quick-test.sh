#!/bin/bash
# Quick test workflow: restore snapshot → sync scripts → boot VM → test
# Usage: ./quick-test.sh [snapshot-name]
# Default snapshot name: clean-install

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_NAME="${1:-clean-install}"
IMAGE="$TEST_DIR/arch-test.qcow2"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

echo "=== Quick Test Workflow ==="
echo ""

# Restore snapshot
echo "[1/2] Restoring to snapshot: $SNAPSHOT_NAME"
qemu-img snapshot -a "$SNAPSHOT_NAME" "$IMAGE"
echo "✓ Restored"
echo ""

# Start VM
echo "[2/2] Starting VM..."
if [ -f "$TEST_DIR/start-qemu-installed.sh" ]; then
    bash "$TEST_DIR/start-qemu-installed.sh"
else
    echo "Error: start-qemu-installed.sh not found"
    exit 1
fi
