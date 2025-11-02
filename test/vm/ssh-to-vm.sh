#!/bin/bash
# SSH into the running Arch ISO VM
# Usage: ./ssh-to-vm.sh [command]

set -eo pipefail

VM_HOST="localhost"
VM_PORT="2222"
VM_USER="root"
VM_PASS="test"

# If command is provided, execute it, otherwise open interactive shell
if [ $# -gt 0 ]; then
    sshpass -p "$VM_PASS" ssh -p "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      "$VM_USER@$VM_HOST" "$@" 2>&1 | grep -v "Warning: Permanently added" || true
else
    sshpass -p "$VM_PASS" ssh -p "$VM_PORT" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      "$VM_USER@$VM_HOST" 2>&1 | grep -v "Warning: Permanently added" || true
fi
