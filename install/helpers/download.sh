#!/bin/bash
# Download all required ArchUp installation files from GitHub

download_section() {
  local title="$1"
  local target_dir="$2"
  shift 2
  local files=("$@")

  for file in "${files[@]}"; do
    curl -sL "$GITHUB_RAW/install/$target_dir/$file" -o "$ARCHUP_INSTALL/$target_dir/$file" 2>/dev/null
  done
  gum style --foreground 2 "[OK] ${title/Downloading /Downloaded }"
}

download_archup_files() {
  echo ""
  gum style --foreground 212 --bold "ArchUp Installer Download"
  echo ""

  local GITHUB_RAW="$ARCHUP_RAW_URL"

  # Create directory structure
  mkdir -p "$ARCHUP_INSTALL"/{helpers,preflight,partitioning,base,config,boot,repos,post-install,post-boot}

  # Core files
  curl -sL "$GITHUB_RAW/install/bootstrap.sh" -o "$ARCHUP_INSTALL/bootstrap.sh" 2>/dev/null
  curl -sL "$GITHUB_RAW/logo.txt" -o "$ARCHUP_PATH/logo.txt" 2>/dev/null
  gum style --foreground 2 "[OK] Core files downloaded"

  # Helpers
  download_section "Downloading helpers..." "helpers" \
    all.sh config.sh logging.sh errors.sh presentation.sh chroot.sh cleanup.sh multilib.sh

  # Preflight
  download_section "Downloading preflight checks..." "preflight" \
    all.sh guards.sh begin.sh identify.sh detect-environment.sh

  # Partitioning
  download_section "Downloading partitioning scripts..." "partitioning" \
    all.sh detect-disk.sh partition.sh format.sh mount.sh

  # Base system
  download_section "Downloading base system scripts..." "base" \
    all.sh kernel.sh enable-multilib.sh cachyos-repo.sh pacstrap.sh fstab.sh

  # Configuration
  download_section "Downloading config scripts..." "config" \
    all.sh system.sh user.sh network.sh zram.sh

  # Bootloader
  download_section "Downloading bootloader setup..." "boot" \
    all.sh limine.sh

  # Repositories
  download_section "Downloading repository setup..." "repos" \
    all.sh multilib.sh aur.sh chaotic.sh

  # Post-install
  download_section "Downloading post-install scripts..." "post-install" \
    all.sh boot-logo.sh plymouth.sh snapper.sh ufw.sh post-boot-setup.sh hooks.sh shell-config.sh verify.sh unmount.sh

  # Post-boot
  download_section "Downloading post-boot scripts..." "post-boot" \
    all.sh snapper.sh ssh-keygen.sh archup-cli.sh archup-first-boot.service

  # Base packages
  curl -sL "$GITHUB_RAW/install/base.packages" -o "$ARCHUP_INSTALL/base.packages" 2>/dev/null
  gum style --foreground 2 "[OK] Package list downloaded"

  echo ""
  gum style --foreground 2 --bold "[OK] All files downloaded successfully"
  echo ""
}
