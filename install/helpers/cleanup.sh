#!/bin/bash
# ArchUp Cleanup - Remove artifacts from failed installations
# This script is 100% idempotent and safe to call multiple times
# Called from bootstrap.sh on every install attempt

# Cleanup modes:
# - default (called from bootstrap): Unmount, close LUKS, disable swap
# - full (optional): Also delete /mnt content, allow fresh pacstrap
# - efi-cleanup (optional): Remove UEFI boot entries (manual only)

set +e  # Continue even if commands fail

CLEANUP_MODE="${1:-default}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

log_cleanup() {
  echo "[CLEANUP] $1"
  if [ -n "${ARCHUP_INSTALL_LOG_FILE:-}" ] && [ -f "$ARCHUP_INSTALL_LOG_FILE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [CLEANUP] $1" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
}

log_success() {
  echo -e "${GREEN}[CLEANUP OK]${NC} $1"
  if [ -n "${ARCHUP_INSTALL_LOG_FILE:-}" ] && [ -f "$ARCHUP_INSTALL_LOG_FILE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [CLEANUP OK] $1" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
}

log_error() {
  echo -e "${RED}[CLEANUP ERROR]${NC} $1"
  if [ -n "${ARCHUP_INSTALL_LOG_FILE:-}" ] && [ -f "$ARCHUP_INSTALL_LOG_FILE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [CLEANUP ERROR] $1" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
}

# ============================================================
# PHASE 1: SAFE UNMOUNTING - Always safe to do
# ============================================================

cleanup_mounts() {
  log_cleanup "Cleaning up mounted filesystems..."

  # Unmount in reverse order (most specific first)
  local mounts_to_clean=(
    "/mnt/boot"
    "/mnt/home"
    "/mnt"
  )

  for mount_point in "${mounts_to_clean[@]}"; do
    if mountpoint -q "$mount_point" 2>/dev/null; then
      log_cleanup "Unmounting $mount_point..."
      if umount -R "$mount_point" 2>/dev/null; then
        log_success "Unmounted $mount_point"
      else
        # Force unmount if normal unmount fails
        log_cleanup "Force unmounting $mount_point..."
        if umount -R -l "$mount_point" 2>/dev/null; then
          log_success "Force unmounted $mount_point"
        else
          log_error "Failed to unmount $mount_point (may retry later)"
        fi
      fi
    fi
  done

  # Wait for any lingering unmount operations
  sleep 1
}

# ============================================================
# PHASE 2: CLOSE ENCRYPTION - Always safe to do
# ============================================================

cleanup_encryption() {
  log_cleanup "Cleaning up LUKS/dm-crypt mappings..."

  # Close cryptroot specifically (Omarchy convention)
  if [ -e /dev/mapper/cryptroot ]; then
    log_cleanup "Closing LUKS container: cryptroot..."
    if cryptsetup close cryptroot 2>/dev/null; then
      log_success "Closed cryptroot"
    else
      log_error "Failed to close cryptroot"
    fi
  fi

  # Close ALL dm-crypt mappings (in case multiple exist)
  local crypt_mappings=$(dmsetup ls --target crypt 2>/dev/null | cut -f1)
  if [ -n "$crypt_mappings" ]; then
    while IFS= read -r mapping; do
      if [ -n "$mapping" ]; then
        log_cleanup "Closing dm-crypt mapping: $mapping..."
        if cryptsetup close "$mapping" 2>/dev/null; then
          log_success "Closed dm-crypt mapping: $mapping"
        else
          log_error "Failed to close $mapping"
        fi
      fi
    done <<< "$crypt_mappings"
  fi

  # Ensure all device mapper entries are removed
  if command -v dmsetup &>/dev/null; then
    if dmsetup table 2>/dev/null | grep -q crypt; then
      log_cleanup "Removing lingering device mapper entries..."
      dmsetup remove_all 2>/dev/null || true
      log_success "Device mapper cleanup attempted"
    fi
  fi
}

# ============================================================
# PHASE 3: DISABLE SWAP - Always safe to do
# ============================================================

cleanup_swap() {
  log_cleanup "Disabling swap..."

  if swapon --show 2>/dev/null | grep -q .; then
    log_cleanup "Found active swap, disabling..."
    if swapoff -a 2>/dev/null; then
      log_success "Swap disabled"
    else
      log_error "Failed to disable swap (non-critical)"
    fi
  fi
}

# ============================================================
# PHASE 4: KILL STUCK PROCESSES - Always safe to do
# ============================================================

cleanup_processes() {
  log_cleanup "Checking for stuck pacstrap/chroot processes..."

  # Kill any stuck pacstrap or arch-chroot processes
  local stuck_pids=$(pgrep -f "pacstrap|arch-chroot" 2>/dev/null)
  if [ -n "$stuck_pids" ]; then
    log_cleanup "Killing stuck processes: $stuck_pids"
    kill $stuck_pids 2>/dev/null || true
    sleep 1
    kill -9 $stuck_pids 2>/dev/null || true
    log_success "Cleaned up stuck processes"
  fi
}

# ============================================================
# PHASE 5: FILESYSTEM SYNC - Always safe to do
# ============================================================

cleanup_sync() {
  log_cleanup "Syncing filesystems..."

  if sync 2>/dev/null; then
    log_success "Filesystems synced"
  else
    log_error "Failed to sync filesystems (non-critical)"
  fi
}

# ============================================================
# OPTIONAL: FULL CLEANUP - Wipe /mnt (use --full flag)
# ============================================================

cleanup_full() {
  log_cleanup "Performing FULL cleanup - deleting /mnt content..."

  if [ -d /mnt ]; then
    log_cleanup "Removing all files and directories in /mnt..."
    if rm -rf /mnt/* 2>/dev/null; then
      log_success "Removed /mnt content"
    else
      # Try alternate approach
      log_cleanup "Retry: Using find to remove /mnt content..."
      find /mnt -mindepth 1 -delete 2>/dev/null || true
      log_success "Removed /mnt content (with find)"
    fi
  fi

  # Ensure /mnt directory exists for next installation
  mkdir -p /mnt 2>/dev/null || true
  log_success "Prepared /mnt for fresh installation"
}

# ============================================================
# OPTIONAL: UEFI BOOT ENTRY CLEANUP (advanced/manual only)
# ============================================================

cleanup_uefi_entries() {
  log_cleanup "Listing existing UEFI boot entries..."
  echo
  echo "Current UEFI boot entries:"
  efibootmgr 2>/dev/null || log_error "efibootmgr not available"
  echo
  echo "To remove archup entries, run:"
  echo "  efibootmgr -b XXXX -B  (where XXXX is the entry number)"
  echo
  log_cleanup "UEFI cleanup requires manual intervention"
}

# ============================================================
# DIAGNOSTIC: Check current state
# ============================================================

cleanup_diagnostic() {
  log_cleanup "=== SYSTEM STATE DIAGNOSTIC ==="

  echo "Mounted filesystems at /mnt:"
  mount | grep /mnt || echo "  (none)"
  echo

  echo "Active dm-crypt mappings:"
  dmsetup ls --target crypt 2>/dev/null || echo "  (none)"
  echo

  echo "Active swap:"
  swapon --show 2>/dev/null || echo "  (none)"
  echo

  echo "Partitions:"
  lsblk 2>/dev/null | head -20
}

# ============================================================
# MAIN EXECUTION
# ============================================================

case "$CLEANUP_MODE" in
  default)
    log_cleanup "=== ArchUp Cleanup (DEFAULT) ==="
    cleanup_mounts
    cleanup_encryption
    cleanup_swap
    cleanup_processes
    cleanup_sync
    log_success "=== Default cleanup complete ==="
    ;;

  full)
    log_cleanup "=== ArchUp Cleanup (FULL - DESTRUCTIVE) ==="
    cleanup_mounts
    cleanup_encryption
    cleanup_swap
    cleanup_processes
    cleanup_full
    cleanup_sync
    log_success "=== Full cleanup complete ==="
    ;;

  uefi-cleanup)
    cleanup_uefi_entries
    ;;

  diagnostic)
    cleanup_diagnostic
    ;;

  *)
    echo "Usage: cleanup.sh [default|full|uefi-cleanup|diagnostic]"
    echo ""
    echo "Modes:"
    echo "  default         - Unmount, close LUKS, disable swap (safe, always done)"
    echo "  full            - Also delete /mnt content (destructive)"
    echo "  uefi-cleanup    - Show UEFI boot entries (manual cleanup)"
    echo "  diagnostic      - Show current system state"
    exit 1
    ;;
esac
