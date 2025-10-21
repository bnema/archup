# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Extra packages**: Separated optional packages into extra.packages installed after Chaotic-AUR
- **Shell verification**: Added shell config verification to post-install verify script
- **Starship config**: Added custom Arch-inspired blue color scheme for starship prompt

### Changed
- **Package organization**: Split packages into base (essential) and extra (optional from Chaotic-AUR)
- **Base system**: Reduced base system to essential packages only, moved modern CLI tools to extra
- **Post-boot error handling**: Improved error handling with non-critical failure messages
- **Shell config structure**: Organized shell files in proper directory structure
- **Plymouth theme**: Changed background color to darker blue (#05142E)
- **Logging**: Removed stdin redirect to fix gum terminal display

### Fixed
- **Shell config ownership**: Fixed chown paths to use chroot-relative paths
- **Chaotic-AUR padding**: Fixed message padding consistency
- **Limine bootloader**: Fixed kernel panic with non-default kernels (cachyos, zen, lts)
- **AMD P-State selection**: Auto-select when only one mode available
- **Bootloader message**: Removed unnecessary Limine message since no choice exists
- **Curl-pipe installation**: Fixed curl-based one-liner installation by auto-detecting piped input, re-executing with TTY for interactive prompts, and running bootstrap before downloads
- **Cleanup flag**: Fixed --cleanup flag to run after helper files are downloaded
- **Verification script**: Disabled ERR trap during verification to properly show all failures and prompt user, use dynamic kernel name instead of hardcoded linux, made fallback initramfs optional

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
