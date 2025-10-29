#!/bin/bash
# Create a full backup copy of the current VM state as a safety net
# Usage: ./backup-golden.sh
# This creates arch-test-golden.qcow2 for disaster recovery

set -e

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="$TEST_DIR/arch-test.qcow2"
BACKUP="$TEST_DIR/arch-test-golden.qcow2"

if [ ! -f "$IMAGE" ]; then
    echo "Error: VM image not found: $IMAGE"
    exit 1
fi

if [ -f "$BACKUP" ]; then
    echo "Golden backup already exists: $BACKUP"
    read -p "Overwrite? (yes/NO): " -r
    if [[ ! $REPLY =~ ^yes$ ]]; then
        echo "Aborted."
        exit 0
    fi
    echo "Removing old backup..."
    rm "$BACKUP"
fi

echo "Creating golden backup..."
echo "This may take a minute..."
cp "$IMAGE" "$BACKUP"

IMAGE_SIZE=$(du -h "$BACKUP" | cut -f1)
echo "✓ Golden backup created: $BACKUP ($IMAGE_SIZE)"
echo ""
echo "To restore from golden backup:"
echo "  cp $BACKUP $IMAGE"
