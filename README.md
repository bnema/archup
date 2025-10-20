<p align="center">
  <img src="assets/archup-logo.svg" alt="ArchUp Logo" width="600">
</p>

## What Is ArchUp?

ArchUp is a minimal, slightly opinionated Arch Linux installer focused on Wayland window managers. It provides a barebone base system with sane defaults and an optional CLI to build your ideal desktop environment.

- **Barebone first** - Install just the essentials and stop here if you want
- **User choice** - Every layer is optional; you decide what to install from a supported list of apps
- **Lightweight focused** - Only Wayland compositors (Niri, Hyprland, Sway, River)
- **Smart foundation** - Critical desktop infrastructure pre-configured (graphics, audio, Wayland, BT, printing)
- **Hardware detection** - Auto-detects GPU and installs correct drivers
- **Coherent UI** - One dark/light theme across all apps (customizable later)
- **Just works** - All infrastructure configured; focus on your workflow, not setup

## What ArchUp Is NOT

- **Not a desktop environment** - We won't support KDE, GNOME, or other full-fledged DEs
- **Not a distro** - It's an installer, not a custom Arch spin
- **Not bloated** - We provide the minimum to get you started
- **Not an Omarchy copy** - We respect Omarchy's architecture but keep it minimal

## Philosophy

ArchUp put the user in control. Every layer is optional:

- Want barebone CLI but with a cool bootloader? Stop at Tier 1
- Need a lightweight GUI? Add Tier 2+3
- Want to customize? Manual changes are easy; we don't lock you in

We learned from Omarchy's excellent modular approach but rejected the bloat. ArchUp provides smart defaults (proven to work) while staying minimal and respecting user choice.

## Quick Start

**Requirements**: AMD/Intel 64-bit system

```bash
# 1. Boot Arch ISO, install barebone system
curl -fsSL https://archup.bnema.dev/install | bash

# 2. Reboot into new system

# 3. Run wizard to add desktop (optional)
archup wizard
```

## Status

**v0.1.0** - Barebone installer complete

- [x] Tier 1: Barebone CLI with LUKS, btrfs, Limine
- [x] Auto-builds `archup` CLI on first boot
- [ ] Tier 2+3: Desktop wizard (in `archup-cli` repo)

**⚠️ Alpha/Testing Phase**: ArchUp is in active development. All contributions, bug reports, and error logs are welcome! Please open an issue if you encounter any problems.

## Links

- **Install**: `https://archup.bnema.dev/install`
- **Dev branch**: `https://archup.bnema.dev/install/dev`
- **CLI repo**: `github.com/bnema/archup-cli`

## Thanks

- **[Omarchy](https://github.com/omakub/omakub)** - For the excellent modular architecture inspiration (MIT License)
- **[Charmbracelet Gum](https://github.com/charmbracelet/gum)** - For the awesome interactive shell toolkit
