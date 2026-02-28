#!/bin/bash
# Push binary to VM for testing
# Usage: ./push-binary-to-vm.sh [--local] /path/to/binary
#
# --local  Also sync local install/ assets/ and logo.txt into /tmp/archup-install/ on
#          the VM and set ARCHUP_LOCAL=1 in /tmp/archup-local-env so the installer
#          skips all GitHub downloads and uses the pre-seeded files instead.

set -eo pipefail

LOCAL_MODE=0
BINARY_PATH=""

for arg in "$@"; do
    case "$arg" in
        --local) LOCAL_MODE=1 ;;
        *)       BINARY_PATH="$arg" ;;
    esac
done

if [ -z "$BINARY_PATH" ]; then
    echo "Usage: $0 [--local] /path/to/binary"
    echo "Example: $0 ../../../dist/archup_linux_amd64_v1/archup"
    echo "         $0 --local ../../../dist/archup_linux_amd64_v1/archup"
    exit 1
fi

BINARY_NAME=$(basename "$BINARY_PATH")

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary file not found: $BINARY_PATH"
    exit 1
fi

VM_HOST="localhost"
VM_PORT="2222"
VM_USER="root"
VM_PASS="test"
REMOTE_PATH="/tmp"
REMOTE_INSTALL_DIR="/tmp/archup-install"

SSH_OPTS="-p $VM_PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
SCP_OPTS="-P $VM_PORT -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
SSHPASS="sshpass -p $VM_PASS"

ssh_cmd() {
    $SSHPASS ssh $SSH_OPTS "$VM_USER@$VM_HOST" "$@" 2>&1 | grep -v "Warning: Permanently added" || true
}

scp_cmd() {
    $SSHPASS scp $SCP_OPTS "$@" 2>&1 | grep -v "Warning: Permanently added" || true
}

rsync_cmd() {
    $SSHPASS rsync -az --no-perms -e "ssh $SSH_OPTS" "$@" 2>&1 | grep -v "Warning: Permanently added" || true
}

echo "Pushing $BINARY_NAME to VM..."

ssh_cmd "mkdir -p $REMOTE_PATH"
scp_cmd "$BINARY_PATH" "$VM_USER@$VM_HOST:$REMOTE_PATH/$BINARY_NAME"

# Verify binary
FILE_SIZE=$(stat -c%s "$BINARY_PATH" 2>/dev/null || stat -f%z "$BINARY_PATH" 2>/dev/null)
REMOTE_SIZE=$(ssh_cmd "stat -c%s $REMOTE_PATH/$BINARY_NAME")

if [ "$FILE_SIZE" = "$REMOTE_SIZE" ]; then
    echo "[OK] Binary pushed successfully!"
    echo "  Location: $REMOTE_PATH/$BINARY_NAME"
    echo "  Size: $FILE_SIZE bytes"
else
    echo "Warning: Size mismatch - transfer may have failed"
    echo "  Local: $FILE_SIZE bytes"
    echo "  Remote: $REMOTE_SIZE bytes"
    exit 1
fi

if [ "$LOCAL_MODE" = "1" ]; then
    # Resolve repo root (two levels up from test/vm/)
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

    echo ""
    echo "Syncing local install files to VM ($REMOTE_INSTALL_DIR)..."

    ssh_cmd "mkdir -p \
        $REMOTE_INSTALL_DIR/configs \
        $REMOTE_INSTALL_DIR/assets/plymouth \
        $REMOTE_INSTALL_DIR/install/configs \
        $REMOTE_INSTALL_DIR/install/mandatory/post-boot"

    # install/ configs (bootstrap files land here: /tmp/archup-install/<file>)
    rsync_cmd "$REPO_ROOT/install/base.packages"                     "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/"
    rsync_cmd "$REPO_ROOT/install/extra.packages"                    "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/"
    rsync_cmd "$REPO_ROOT/install/configs/limine.conf.template"      "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/configs/"
    rsync_cmd "$REPO_ROOT/install/configs/chaotic-aur.conf"          "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/configs/"
    rsync_cmd "$REPO_ROOT/install/configs/limine-update.hook"        "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/configs/"

    # install/ configs also at repo-relative path (used by tryReadLocal fallback)
    rsync_cmd "$REPO_ROOT/install/configs/limine.conf.template"      "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/install/configs/"
    rsync_cmd "$REPO_ROOT/install/configs/limine-update.hook"        "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/install/configs/"

    # assets
    rsync_cmd "$REPO_ROOT/assets/plymouth/"                          "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/assets/plymouth/"

    # logo.txt
    rsync_cmd "$REPO_ROOT/logo.txt"                                  "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/"

    # post-boot scripts
    rsync_cmd "$REPO_ROOT/install/mandatory/post-boot/"              "$VM_USER@$VM_HOST:$REMOTE_INSTALL_DIR/install/mandatory/post-boot/"

    # Write env file so the installer wrapper can source ARCHUP_LOCAL=1
    ssh_cmd "echo 'export ARCHUP_LOCAL=1' > /tmp/archup-local-env"

    echo "[OK] Local files synced to $REMOTE_INSTALL_DIR"
    echo "[OK] /tmp/archup-local-env written — run: source /tmp/archup-local-env && $REMOTE_PATH/$BINARY_NAME"
fi
