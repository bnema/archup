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

## Testing Workflow

```bash
# Sync local changes to VM
./sync-scripts-to-vm.sh

# SSH into VM
ssh -p 2222 root@localhost

# Run installer
cd ~/.local/share/archup && ./install.sh
```

## Scripts

- `create-vm-img.sh` - Create disk image
- `start-qemu-arch-iso.sh` - Boot ISO (auto-downloads)
- `start-qemu-installed.sh` - Boot installed system
- `sync-scripts-to-vm.sh` - Upload changes to VM

## Notes

- SSH: `localhost:2222` → VM `:22`
- ISO/qcow2 files gitignored
