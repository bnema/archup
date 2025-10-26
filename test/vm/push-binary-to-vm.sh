#!/bin/bash
# Push binary to VM for testing
# Usage: ./push-binary-to-vm.sh /path/to/binary

set -eo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: $0 /path/to/binary"
    echo "Example: $0 ../../../dist/archup-installer_linux_amd64_v1/archup-installer"
    exit 1
fi

BINARY_PATH="$1"
BINARY_NAME=$(basename "$BINARY_PATH")

# Check if file exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary file not found: $BINARY_PATH"
    exit 1
fi

VM_HOST="localhost"
VM_PORT="2222"
VM_USER="root"
VM_PASS="test"
REMOTE_PATH="/root/.local/share/archup"

echo "Pushing $BINARY_NAME to VM..."

# Create remote directory
sshpass -p "$VM_PASS" ssh -p "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "$VM_USER@$VM_HOST" "mkdir -p $REMOTE_PATH" 2>&1 | grep -v "Warning: Permanently added" || true

# Send binary
sshpass -p "$VM_PASS" scp -P "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "$BINARY_PATH" "$VM_USER@$VM_HOST:$REMOTE_PATH/$BINARY_NAME" 2>&1 | grep -v "Warning: Permanently added" || true

# Verify
FILE_SIZE=$(stat -c%s "$BINARY_PATH" 2>/dev/null || stat -f%z "$BINARY_PATH" 2>/dev/null)
REMOTE_SIZE=$(sshpass -p "$VM_PASS" ssh -p "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "$VM_USER@$VM_HOST" "stat -c%s $REMOTE_PATH/$BINARY_NAME" 2>&1 | grep -v "Warning: Permanently added" || true)

if [ "$FILE_SIZE" = "$REMOTE_SIZE" ]; then
    echo "✓ Binary pushed successfully!"
    echo "  Location: $REMOTE_PATH/$BINARY_NAME"
    echo "  Size: $FILE_SIZE bytes"
else
    echo "⚠ Size mismatch - transfer may have failed"
    echo "  Local: $FILE_SIZE bytes"
    echo "  Remote: $REMOTE_SIZE bytes"
    exit 1
fi
