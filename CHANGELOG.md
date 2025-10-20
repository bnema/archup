# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- **README**: Added ArchUp logo and refined project description for clarity
- **Kernel selection**: Simplified AMD P-State mode descriptions for better user clarity

### Fixed
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
