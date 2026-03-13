<p align="center">
  <img src="assets/archup-logo-v2-5.png" alt="ArchUp Logo" width="600">
</p>

## What Is ArchUp?

ArchUp is an Arch Linux installer similar to archinstall but it makes the choices for you on the boring parts. The difference with Omarchy is that it's not enforcing any dotfile or app. The goal is to install Arch as fast as possible with sane defaults.

**What it decides for you:**
- Btrfs with `@` and `@home` subvolumes
- Limine bootloader (UEFI only)
- Chaotic-AUR enabled out of the box
- Plymouth boot splash
- Snapper for snapshot-based rollbacks
- NetworkManager, OpenSSH, zram, firewalld

**What you choose:**
- Disk and optional LUKS2 encryption
- Hostname, user, locale, timezone, keymap
- Kernel (linux, linux-lts, linux-zen, linux-hardened, linux-cachyos)
- AMD P-State mode (auto-detected per Zen generation)
- GPU drivers (auto-detected)
- Extra repos (CachyOS, AUR helper)
- Dank Linux desktop on first boot (optional)

## Get Started

**Requirements:** x86\_64, UEFI, Secure Boot disabled

Boot the Arch ISO and run:

```bash
curl -fsSL https://archup.run/install | bash
```

Reboot when done.

## What's Installed

**Base system:**
- Btrfs filesystem, Limine bootloader, Plymouth
- Kernel of your choice + matching microcode
- GPU drivers and firmware (auto-detected)
- NetworkManager, OpenSSH, systemd-resolved, zram

**CLI tools:**
- Neovim, Git, sudo, man pages
- Modern utilities: fzf, ripgrep, bat, eza, zoxide, starship, btop, yazi
- Build tools: base-devel, Go, Rust (for AUR)

**First-boot setup (automatic):**
- Snapper configured for Btrfs snapshots
- firewalld enabled
- SSH host keys generated
- ble.sh and shell tooling installed

## Status

Active development — tested manually on real hardware.

- [x] TUI installer (functional)
- [x] LUKS2 encryption support
- [x] First-boot systemd service
- [ ] Dank Linux desktop (niri/Hyprland) — in progress

Report bugs with logs from `/var/log/archup.log`.

## Acknowledgments

- [Omarchy](https://github.com/omakub/omakub) — modular architecture approach
- [Charmbracelet](https://github.com/charmbracelet) — TUI toolkit (Bubble Tea, Lipgloss, Huh)
