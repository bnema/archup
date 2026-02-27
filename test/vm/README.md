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

## Hardware Profiles

The QEMU launchers support two hardware profiles that simulate realistic PC hardware. Both use `q35` machine type, NVMe storage, Intel e1000e NIC, XHCI USB, Intel HDA audio, and virtio-vga GPU.

| Profile | System | Board |
|---------|--------|-------|
| `desktop` (default) | Gigabyte B550 AORUS ELITE V2 | Gigabyte B550 |
| `laptop` | ThinkPad X1 Carbon Gen 10 | LENOVO 21CB |

### Usage

```bash
# Desktop profile (default)
./start-qemu-installed.sh
./start-qemu-installed.sh --profile desktop
./start-qemu-desktop.sh                      # convenience wrapper

# Laptop profile
./start-qemu-installed.sh --profile laptop
./start-qemu-laptop.sh                       # convenience wrapper

# Secure Boot (ArchUp preflight expects it OFF by default)
./start-qemu-installed.sh --secure-boot on

# ISO boot with profile
./start-qemu-arch-iso.sh --profile laptop

# Custom SSH port
./start-qemu-installed.sh --ssh-port 2223
```

### Verify Hardware Inside the VM

After booting, run these inside the VM to confirm the hardware identity:

```bash
# BIOS/system/board identity (should show Gigabyte or LENOVO strings)
dmidecode -t 0,1,2,3

# PCI devices (check for virtio-vga, e1000e)
lspci -nn

# Storage identity (should show nvme0n1 with serial TESTDESKTOP0001 or TESTLAPTOP00001)
lsblk -o NAME,MODEL,SERIAL,TYPE

# UEFI boot entries
bootctl status

# NIC name (e1000e presents as enp*s0 or similar)
ip link
```

### Notes

- Secure Boot mode copies OVMF_VARS to a temp file each run; keys must be enrolled manually via UEFI shell on first boot.
- `check-luks.sh` defaults to `/dev/nvme0n1p2`. For old virtio-disk images use `--disk /dev/vda2`.
- `arch-test.qcow2` must exist before launching any VM (run `./create-vm-img.sh` first). The ISO launcher also requires it as the install target disk.
- Desktop and laptop profiles use different MAC addresses (`52:54:00:12:34:01` / `:02`) so both VMs can run concurrently on the same host.
