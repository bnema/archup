# VM Testing

Quick QEMU testing environment for ArchUp installer.

## Prerequisites

```bash
sudo pacman -S qemu-full edk2-ovmf sshpass rsync
```

## Setup

```bash
# 1. Create disk image
./create-vm-img.sh

# 2. Start VM with Arch ISO
./start-qemu-arch-iso.sh

# 3. In the VM, start SSH and set password
systemctl start sshd
passwd  # Set password to: test

# 4. From host, setup SSH keys
ssh-copy-id -p 2222 root@localhost  # Use password: test
```

## Quick Test Cycle

```bash
# One-time: after install, create snapshot
./snapshot-create.sh clean-install

# Iterate: reset → test → poweroff → repeat
./snapshot-restore.sh clean-install
./start-qemu-installed.sh
```

## Scripts

- `snapshot-create.sh [name]` - Create snapshot
- `snapshot-restore.sh [name]` - Restore snapshot
- `snapshot-list.sh` - List snapshots
- `snapshot-delete.sh <name>` - Delete snapshot
- `quick-test.sh [name]` - Restore + boot
- `backup-golden.sh` - Full backup for safety

## Notes

- SSH: `localhost:2222` → VM `:22`
- ISO/qcow2 files gitignored
