# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Separate encryption password**: Allow users to choose between using their account password or a separate password for disk encryption during setup

## [0.5.1] - 2026-03-13

### Fixed
- **Fallback initramfs warning**: `configureMkinitcpio` now checks for `initramfs-<kernel>-fallback.img` after `mkinitcpio -P` and emits a structured warning if absent, surfacing the root cause of `limine-snapper-notify` failures on first boot
- **jq added to extra packages**: `jq` is now installed as part of the default extra package set

### Removed
- **archup-cli post-boot install**: Removed `archup-cli.sh` and its invocation from `all.sh`; the post-boot model is barebone or Dank Linux only

## [0.5.0] - 2026-03-13

### Added
- **Go CLI binary**: Rewrote the installer as a compiled Go binary using Cobra, Bubble Tea, Bubbles, Huh, and Lipgloss; replaced the bash orchestrator with a Go binary downloader (`install/bootstrap.sh`)
- **DDD architecture**: Introduced `internal/domain`, `internal/application`, `internal/infrastructure`, and `internal/interfaces` layers with domain entities, ports, and adapters
- **Kernel selection**: Added TUI screen for choosing between linux, linux-lts, linux-zen, and linux-cachyos kernels
- **GPU detection**: Added GPU detection and automatic driver/package selection during bootstrap
- **Disk encryption toggle**: Added TUI screen for enabling LUKS encryption with separate password support
- **Dank Linux opt-in**: Added toggle screen before installation and first-boot auto-run via flag file
- **AUR helper and Chaotic-AUR selection**: Added TUI screen; Chaotic AUR is always enabled; user selects helper (paru/yay)
- **CachyOS repository support**: Added host-side and chroot-side CachyOS repo setup when linux-cachyos kernel is selected
- **Limine bootloader**: Replaced systemd-boot with Limine; added nested snapshot entries (`//Snapshots`), machine-id injection, and a pacman hook for kernel upgrades
- **limine-snapper-sync**: Install and enable the limine-snapper-sync service for automatic snapshot boot entries
- **Plymouth theme**: Added custom theme with ASCII art PNG logo, centered layout, scaled LUKS password entry, and black background
- **Power management**: Added zram-generator for compressed swap and enabled power-profiles-daemon
- **Post-boot first-login hook**: Moved DMS opt-in and firewalld configuration to first-login; deploy Starship prompt
- **Installation verification gate**: Added post-install verification with structured warnings before completing
- **AMD P-State configuration**: Added comprehensive AMD CPU detection and P-State mode selection
- **WiFi credential migration**: Migrate iwd credentials to NetworkManager profiles post-install
- **bleu-theme integration**: Integrated bleu-theme for CLI tool theming
- **ble.sh**: Added Bash Line Editor installation on first boot
- **Wayland runtime libs**: Added essential runtime libraries for Wayland desktop sessions
- **Nerd fonts**: Added ttf-ibmplex-mono-nerd and ttf-firacode-nerd to extra packages
- **nvtop**: Added nvtop GPU monitor to extra packages
- **Audio stack**: Added Pipewire (pipewire, pipewire-alsa, pipewire-jack, pipewire-pulse, wireplumber), pamixer, playerctl, and gst plugins
- **Bluetooth utilities**: Added bluetui alongside bluez stack
- **WiFi packages**: Added wireless-regdb, impala, and broadcom-wl-dkms
- **firewalld**: Replaced ufw with firewalld; added post-boot initialization script
- **Local debug mode**: Added flag to skip GitHub downloads during VM testing
- **Comprehensive unit tests**: Added test suites for preflight, bootstrap, base install, config, and system phases; centralized test helpers

### Changed
- **Installer entrypoint**: `cmd/archup` replaces the former bash entry point; build managed with goreleaser and GNU Make
- **TUI architecture**: Converted to thin interface layer with flat section navigation, unified keyboard hints, and dedicated view files
- **Shell configuration**: Stripped to no-op; all shell setup now handled by `cli-tools.sh` on first boot
- **Post-boot scripts**: Refactored into `PostBootScripts` list; removed `copyShellConfigs`; added TTY and network ordering to systemd service
- **Chaotic-AUR setup order**: Keyring and mirrorlist installed before adding to `pacman.conf` to prevent pacman errors
- **Branding**: Renamed UEFI boot label and Limine branding from "ArchUp" to "Arch Linux"; updated installer title with version, default hostname `arch`, and black background

### Removed
- **Bash orchestrator scripts**: Removed shell scripts superseded by Go handlers (DDD migration cleanup)
- **systemd-boot**: Consolidated entirely on Limine bootloader
- **Legacy UI and phases code**: Removed dead `internal/ui` and `internal/phases` code paths from pre-DDD era
- **UFW**: Replaced by firewalld

### Fixed
- **Security**: Prevented password exposure in process listings and config files
- **NetworkManager**: Stopped masking `NetworkManager-wait-online` to restore DNS readiness; added `nss-lookup.target` ordering and DNS readiness gate
- **Bootloader**: Corrected Limine path/cmdline keys, nested `/+` format, default entry, and quiet mode flags
- **Plymouth**: Fixed centering, scaling, bullet positions, background color, and lock icon removal
- **Post-boot**: Fixed bash_profile self-cleanup placement, HOME-relative log path, TTY log bleed suppression
- **CachyOS**: Added `pacman-key --init` and `MkdirAll` before CachyOS repo setup
- **AUR helpers**: Built from source as non-root user; handle Chaotic-AUR unavailability; used `makepkg --packagelist` for deterministic installation
- **Chaotic-AUR mirrors**: Fetch dynamically from official mirrorlist with fallback support
- **Logging**: Added comprehensive UI event logging; prevented stdout leak to TUI
- **CodeRabbit review findings**: Fixed `PrevSection` logic, `Branch()` empty guard, `isEncrypted` dedup, `DISK_PLACEHOLDER` constant, regex pacman uncomment, UEFI-safe limine hook

## [0.3.0] - 2025-10-22

### Added
- **UFW firewall**: Added ufw with default deny incoming policy (configured on first boot)
- **First-boot branding**: Display ArchUp logo and completion message during first-boot setup
- **Pacman configuration**: Enable color output, parallel downloads, and ILoveCandy progress bar on installed system
- **Fast ISO downloads**: Configure ISO pacman with ParallelDownloads=10 before base installation
- **zram swap**: Configured zram-generator for compressed in-memory swap with optimized sysctl parameters
- **Broadcom WiFi**: Added broadcom-wl driver to base packages for Broadcom wireless support
- **Multilib support**: Enabled multilib repository on ISO and installed system with reusable helper

### Fixed
- **Download script**: Added missing extra.packages to download list
- **UFW setup**: Moved firewall configuration to post-boot to avoid kernel module errors in chroot
- **Spinner visibility**: Fixed gum spinner not showing during logged phase by redirecting stderr to /dev/tty

### Changed
- **Disk selection**: Enhanced disk listing with model, serial, and vendor information using JSON output
- **Install URL**: Updated to archup.run domain with /install and /i shortcuts
- **archup-cli installation**: Switched to go install for user-owned binary and easier self-updates
- **Network manager**: Replaced systemd-networkd + iwd with NetworkManager for easier WiFi management via nmcli/nmtui
- **Shell aliases**: Added fd as find replacement alias

### Fixed
- **Help flag**: Moved --help check before any side effects for proper argument parsing
- **Error handling**: Fixed ERR trap propagation with set -E and proper cleanup execution
- **Spinner output**: Fixed package installation output leaking through spinner display
- **Cleanup flag**: Fixed --cleanup flag triggering false errors from pgrep exit codes

## [0.2.0] - 2025-10-21

### Added
- **Extra packages**: Separated optional packages into extra.packages installed after Chaotic-AUR
- **Shell verification**: Added shell config verification to post-install verify script
- **Starship config**: Added custom Arch-inspired blue color scheme for starship prompt
- **System utilities**: Added essential tools (dosfstools, which, wget, unzip, p7zip) to extra packages
- **Bluetooth support**: Added bluez and bluez-utils for Bluetooth management
- **OpenSSH**: Added openssh to base packages for ssh-keygen and SSH client tools

### Changed
- **Package organization**: Split packages into base (essential) and extra (optional from Chaotic-AUR)
- **Base system**: Reduced base system to essential packages only, moved modern CLI tools to extra
- **Post-boot error handling**: Improved error handling with non-critical failure messages
- **Shell config structure**: Organized shell files in proper directory structure
- **Plymouth theme**: Changed background color to darker blue (#050a14)
- **Logging**: Removed stdin redirect to fix gum terminal display
- **Asset URLs**: Changed ARCHUP_RAW_URL from dev to main branch to prevent breakage after merge

### Fixed
- **Shell config ownership**: Fixed chown paths to use chroot-relative paths
- **Chaotic-AUR padding**: Fixed message padding consistency
- **Limine bootloader**: Fixed kernel panic with non-default kernels (cachyos, zen, lts)
- **AMD P-State selection**: Auto-select when only one mode available
- **Bootloader message**: Removed unnecessary Limine message since no choice exists
- **Curl-pipe installation**: Fixed curl-based one-liner installation by auto-detecting piped input, re-executing with TTY for interactive prompts, and running bootstrap before downloads
- **Cleanup flag**: Fixed --cleanup flag to run after helper files are downloaded
- **Verification script**: Disabled ERR trap during verification to properly show all failures and prompt user, use dynamic kernel name instead of hardcoded linux, made fallback initramfs optional
- **Package name**: Fixed git-delta package name (was incorrectly named 'delta')
- **SSH key generation**: Fixed post-boot SSH key generation by including openssh in base packages

## [0.1.0] - 2025-10-21

### Added
- **Post-boot setup**: Dedicated post-boot setup script for first-boot execution
- **archup-cli integration**: Added archup-cli installation to first-boot sequence
- **archup-cli builder**: Download and install archup-cli from latest GitHub release during installation

### Changed
- **Helper script refactoring**: Updated helper scripts for improved architecture
- **Build system**: Updated Makefile to remove preset paths and add post-boot paths
- **AUR helpers**: Consolidated AUR helpers into single unified aur.sh script
- **Installer simplification**: Simplified install.sh to use download helper and remove redundant preset logic

### Fixed
- Post-install workflow: Removed duplicate post-boot setup logic from snapper script
- Installation sequence: Properly integrated post-boot-setup into main installation flow
