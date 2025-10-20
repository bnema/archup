#!/bin/bash
# Quick sync script to upload all install files to VM for testing
# Uses rsync for fast synchronization
# Usage: ./sync-vm.sh

set -eo pipefail

VM_HOST="localhost"
VM_PORT="2222"
VM_USER="root"
VM_PASS="test"
LOCAL_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REMOTE_PATH="/root/.local/share/archup"

echo "Syncing ArchUp install files to VM..."

# Create remote directory first
sshpass -p "$VM_PASS" ssh -p "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  "$VM_USER@$VM_HOST" "mkdir -p $REMOTE_PATH" 2>&1 | grep -v "Warning: Permanently added" || true

# Sync only install directory and root scripts
sshpass -p "$VM_PASS" rsync -az --progress \
  -e "ssh -p $VM_PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
  "$LOCAL_PATH/install/" \
  "$VM_USER@$VM_HOST:$REMOTE_PATH/install/" 2>&1 | grep -v "Warning: Permanently added" || true

sshpass -p "$VM_PASS" rsync -az --progress \
  -e "ssh -p $VM_PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
  "$LOCAL_PATH"/*.sh "$LOCAL_PATH"/logo.txt \
  "$VM_USER@$VM_HOST:$REMOTE_PATH/" 2>&1 | grep -v "Warning: Permanently added" || true

echo
echo "✓ Sync complete!"
