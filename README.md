# archup

**A fast, minimal Arch Linux auto-installer alternative to archinstall**

## Project Status

✅ **Phase 0: Core Infrastructure - COMPLETE**
✅ **Phase 1: Barebone Installer - Basic - COMPLETE**
✅ **Phase 2: Add btrfs + LUKS Encryption - COMPLETE**

### Completed
- ✅ Project structure created
- ✅ Helper utilities implemented (logging, errors, presentation, chroot)
- ✅ Main orchestrator (install.sh)
- ✅ Barebone package list (~15 packages)
- ✅ Logo designed
- ✅ System guards (vanilla Arch, UEFI, x86_64, Secure Boot checks)
- ✅ Bootloader selection (systemd-boot or limine)
- ✅ Auto-partitioning (GPT, EFI + root)
- ✅ Base system installation (pacstrap + fstab)
- ✅ System configuration (timezone, locale, hostname, user creation)
- ✅ Network configuration (systemd-networkd + iwd)
- ✅ Bootloader installation (systemd-boot with boot entries)
- ✅ btrfs filesystem with subvolumes (@ for root, @home for home)
- ✅ LUKS encryption with Argon2id (optional, 2000ms iteration)
- ✅ Encrypted boot support (mkinitcpio hooks)

### Project Structure
```
archup/
├── README.md
├── logo.txt
├── install.sh                 # Main orchestrator
├── install/
│   ├── helpers/
│   │   ├── all.sh            # Source all helpers
│   │   ├── logging.sh        # Log file + colorized output
│   │   ├── errors.sh         # Trap handlers, error display
│   │   ├── presentation.sh   # UI helpers (cursor, colors)
│   │   └── chroot.sh         # Chroot utilities
│   ├── preflight/            # Phase 1 ✅
│   │   ├── all.sh            # Preflight orchestrator
│   │   ├── guards.sh         # System validation
│   │   └── begin.sh          # Welcome and bootloader selection
│   ├── partitioning/         # Phase 1 ✅
│   │   ├── all.sh            # Partitioning orchestrator
│   │   ├── detect-disk.sh    # Disk selection
│   │   ├── partition.sh      # GPT partitioning (EFI + root)
│   │   ├── format.sh         # Format as FAT32 + ext4
│   │   └── mount.sh          # Mount partitions to /mnt
│   ├── base/                 # Phase 1 ✅
│   │   ├── all.sh            # Base installation orchestrator
│   │   ├── pacstrap.sh       # Install base packages
│   │   └── fstab.sh          # Generate fstab
│   ├── config/               # Phase 1 ✅
│   │   ├── all.sh            # Configuration orchestrator
│   │   ├── system.sh         # Timezone, locale, hostname
│   │   ├── user.sh           # User creation and sudo setup
│   │   └── network.sh        # Network configuration
│   ├── boot/                 # Phase 1 ✅
│   │   ├── all.sh            # Bootloader orchestrator
│   │   └── systemd-boot.sh   # systemd-boot installation
│   ├── security/             # (Future - TODO)
│   ├── post-install/         # (Future - TODO)
│   └── presets/
│       └── barebone.packages # Minimal package list (~15 packages)
├── bin/                      # (Future utilities)
└── docs/                     # (Future documentation)
```

## Installation

### Quick Install (from Arch ISO)

```bash
# Download and run archup installer
wget -O- https://raw.githubusercontent.com/bnema/archup/main/install.sh | bash
```

Or download first, then run:

```bash
# Download the installer
wget https://raw.githubusercontent.com/bnema/archup/main/install.sh

# Make it executable
chmod +x install.sh

# Run the installer
./install.sh
```

**Note:** Phase 1 complete! The installer can now install a minimal bootable Arch Linux system. Ready for VM testing.

---

## Development

### Testing Syntax
```bash
# Check all shell scripts for syntax errors and linting issues
make check

# Check syntax only (bash -n)
make check-syntax

# Run shellcheck linting
make check-shellcheck
```

### Testing Installation (Phase 1)
To test the installer in a VM (QEMU/KVM recommended):

```bash
# Boot Arch ISO in VM with UEFI enabled
# Connect to internet
# Download archup
git clone https://github.com/bnema/archup.git
cd archup

# Set the archup path
export ARCHUP_PATH=$(pwd)

# Run the installer
sudo ./install.sh
```

**What the installer does:**
1. Validates system (vanilla Arch ISO, UEFI, x86_64)
2. Asks for bootloader preference (systemd-boot or limine)
3. Asks for encryption preference (optional LUKS with Argon2id)
4. Selects installation disk (auto-partitions with GPT)
5. Formats partitions (FAT32 EFI + btrfs root with optional LUKS)
6. Creates btrfs subvolumes (@ for root, @home for home)
7. Installs ~15 base packages with pacstrap (+ cryptsetup if encrypted)
8. Configures system (timezone, locale, hostname, user)
9. Installs systemd-boot bootloader (with encryption support if enabled)
10. Creates bootable minimal Arch system

## Environment Variables

- `ARCHUP_PATH` - Installation directory (default: `$HOME/.local/share/archup`)
- `ARCHUP_REPO_URL` - Repository URL for bug reports (default: `https://github.com/bnema/archup`)
- `ARCHUP_CHROOT_INSTALL` - Set to `1` for chroot mode

## Next Steps

**Phase 3: Add Limine Bootloader Option** (Week 3)
- [ ] Support Limine as alternative to systemd-boot
- [ ] Create install/boot/limine.sh script
- [ ] Configure limine.conf for UEFI
- [ ] Add encryption support for Limine
- [ ] Test both bootloaders with encryption

## Reference

Inspired by:
- **Omarchy installer** - Modular architecture and security implementations
- **archinstall** - Official Arch installer
- **Arch philosophy** - Simplicity, minimalism, user choice

## License

TBD