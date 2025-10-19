# archup

**A fast, minimal Arch Linux auto-installer alternative to archinstall**

## Project Status

✅ **Phase 0: Core Infrastructure - COMPLETE**
✅ **Phase 1: Barebone Installer - Basic - COMPLETE**
✅ **Phase 2: Add btrfs + LUKS Encryption - COMPLETE**
✅ **Phase 3: Limine Bootloader - COMPLETE**
✅ **Phase 4: Kernel Selection + Microcode - COMPLETE**
✅ **Phase 5: Repository Setup (AUR + Chaotic) - COMPLETE**
✅ **Phase 6: Barebone Preset Complete - COMPLETE**

### Completed
- ✅ Project structure created
- ✅ Helper utilities implemented (logging, errors, presentation, chroot)
- ✅ Main orchestrator (install.sh)
- ✅ Barebone package list (~15 packages)
- ✅ Logo designed
- ✅ System guards (vanilla Arch, UEFI, x86_64, Secure Boot checks)
- ✅ Limine bootloader (exclusive - superior btrfs support)
- ✅ Auto-partitioning (GPT, EFI + root)
- ✅ Base system installation (pacstrap + fstab)
- ✅ System configuration (timezone, locale, hostname, user creation)
- ✅ Network configuration (systemd-networkd + iwd)
- ✅ Limine bootloader installation with UEFI boot entries
- ✅ btrfs filesystem with subvolumes (@ for root, @home for home)
- ✅ LUKS encryption with Argon2id (optional, 2000ms iteration)
- ✅ Encrypted boot support (mkinitcpio hooks)
- ✅ Kernel selection (linux, linux-lts, linux-zen)
- ✅ Auto-detected CPU microcode (intel-ucode or amd-ucode)
- ✅ AMD P-State driver selection (active/guided/passive)
- ✅ Optional AUR support (yay helper installation)
- ✅ Optional Chaotic-AUR repository
- ✅ Keyboard layout migration from ISO
- ✅ WiFi credentials migration from ISO (iwd)

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
# Download the installation script (curl is available by default on Arch ISO)
curl -L https://raw.githubusercontent.com/bnema/archup/main/install.sh -o install.sh
chmod +x install.sh

# Run the installer (it will download all required files)
sudo ./install.sh
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

### Testing Installation
To test the installer in a VM (QEMU/KVM recommended):

```bash
# Boot Arch ISO in VM with UEFI enabled
# Connect to internet

# Download and run installer (automatically downloads all required files)
curl -L https://raw.githubusercontent.com/bnema/archup/main/install.sh -o install.sh
chmod +x install.sh
sudo ./install.sh
```

**What the installer does:**
1. Validates system (vanilla Arch ISO, UEFI, x86_64)
2. Asks for encryption preference (optional LUKS with Argon2id)
3. Asks for kernel choice (linux, linux-lts, or linux-zen)
4. Auto-detects CPU and selects microcode (Intel/AMD)
5. For AMD: Asks for P-State driver mode (active/guided/passive)
6. Selects installation disk (auto-partitions with GPT)
7. Formats partitions (FAT32 EFI + btrfs root with optional LUKS)
8. Creates btrfs subvolumes (@ for root, @home for home)
9. Installs base packages with pacstrap (kernel + microcode + cryptsetup if encrypted)
10. Configures system (timezone, locale, hostname, user)
11. Installs Limine bootloader (with encryption and AMD tuning if enabled)
12. Optionally installs yay AUR helper (if user enables AUR support)
13. Optionally adds Chaotic-AUR repository (pre-built AUR packages)
14. Creates bootable minimal Arch system with btrfs

## Environment Variables

- `ARCHUP_PATH` - Installation directory (default: `$HOME/.local/share/archup`)
- `ARCHUP_REPO_URL` - Repository URL for bug reports (default: `https://github.com/bnema/archup`)
- `ARCHUP_CHROOT_INSTALL` - Set to `1` for chroot mode

## Next Steps

**🎉 Milestone 1 Reached!** - Barebone preset is production-ready and VM-testable

**Phase 7: Default Preset - Minimal GUI** (Week 5)
- [ ] Add niri desktop environment
- [ ] Setup SDDM display manager
- [ ] Add minimal Wayland apps (waybar, mako, fuzzel)
- [ ] Add kitty terminal and Firefox browser

## Reference

Inspired by:
- **Omarchy installer** - Modular architecture and security implementations
- **archinstall** - Official Arch installer
- **Arch philosophy** - Simplicity, minimalism, user choice

## License

TBD