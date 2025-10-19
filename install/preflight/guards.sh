#!/bin/bash
# Preflight guards for archup installer
# Validates system requirements before installation

abort() {
  gum style --foreground 1 --padding "1 0 0 $PADDING_LEFT" "archup install requires: $1"
  echo
  echo "Preflight check failed: $1"
  gum confirm "Proceed anyway on your own accord and without assistance?" --padding "0 0 0 $PADDING_LEFT" || exit 1
}

gum style --foreground 4 --padding "1 0 0 $PADDING_LEFT" "Running preflight checks..."

# Must be running from Arch ISO or vanilla Arch
if [[ ! -f /etc/arch-release ]]; then
  abort "Vanilla Arch Linux or Arch ISO"
fi

# Must not be an Arch derivative distro
for marker in /etc/cachyos-release /etc/eos-release /etc/garuda-release /etc/manjaro-release; do
  if [[ -f "$marker" ]]; then
    abort "Vanilla Arch (detected derivative distro)"
  fi
done

# Must be x86_64 architecture
if [ "$(uname -m)" != "x86_64" ]; then
  abort "x86_64 CPU architecture"
fi

# Must have UEFI boot (check for EFI variables)
if [[ ! -d /sys/firmware/efi/efivars ]]; then
  abort "UEFI boot mode (legacy BIOS not supported)"
fi

# Must have Secure Boot disabled
if bootctl status 2>/dev/null | grep -q 'Secure Boot: enabled'; then
  abort "Secure Boot disabled"
fi

# Cleared all guards
gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "All preflight checks passed"
