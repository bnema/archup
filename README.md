<p align="center">
  <img src="assets/archup-logo-v2-5.png" alt="ArchUp Logo" width="600">
</p>

ArchUp is an opinionated Arch Linux installer that picks the boring defaults so you don't have to.

## What It Decides

- Btrfs with `@` and `@home` subvolumes
- Limine bootloader (UEFI only)
- Chaotic-AUR enabled
- Plymouth boot splash
- Snapper for snapshot-based rollbacks
- NetworkManager, OpenSSH, zram, firewalld

## What You Choose

- Disk and optional LUKS2 encryption
- Hostname, user, locale, timezone, keymap
- Kernel: `linux`, `linux-lts`, `linux-zen`, `linux-hardened`, `linux-cachyos`
- AMD P-State mode (auto-detected per Zen generation)
- GPU drivers (auto-detected)
- Extra repos (CachyOS, AUR helper)
- Dank Linux desktop on first boot (optional)

## Quick Start

**Requirements:** x86_64, UEFI, Secure Boot disabled

Boot the Arch ISO and run:

```bash
curl -fsSL https://archup.run/install | bash
```

Reboot when done.

## What Gets Installed

**Base packages:** `base`, `linux-firmware`, `btrfs-progs`, `limine`, `efibootmgr`, `plymouth`, `networkmanager`, `openssh`, `firewalld`, `sudo`, `neovim`, `man-db`, `git`, `zram-generator`, Fira Sans, JetBrains Mono Nerd Font

**Extra packages:** `snapper`, `snap-pac`, `base-devel`, Go, Rust/Cargo, GTK3/4/libadwaita, Qt5/6 + Wayland, GStreamer, ffmpeg, Pipewire, Bluetooth (bluez), WiFi (iwd), modern CLI tools (`eza`, `zoxide`, `fzf`, `ripgrep`, `bat`, `fd`, `btop`, `nvtop`, `yazi`, `starship`, `jq`, `git-delta`, `dust`, `duf`, `tealdeer`, `tree`)

**First-boot service (automatic):**
- Snapper configured for Btrfs snapshots
- firewalld enabled
- SSH host keys generated
- ble.sh and shell tooling installed
- Optional display manager opt-in

## Status

Active development, tested on real hardware.

- [x] TUI installer
- [x] LUKS2 encryption
- [x] First-boot systemd service
- [ ] Dank Linux desktop (niri/Hyprland), in progress

Report bugs with logs from `/var/log/archup.log`.

## Acknowledgments

- [Omarchy](https://github.com/omakub/omakub) for the modular architecture approach
- [Charmbracelet](https://github.com/charmbracelet) for the TUI toolkit (Bubble Tea, Lipgloss, Huh)
