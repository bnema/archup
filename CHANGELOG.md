# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

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
