#!/bin/bash
# Shared QEMU argument builder for hardware profiles
# Source this file then call: build_qemu_args PROFILE SECURE_BOOT SSH_PORT TEST_DIR
# Result is stored in the QEMU_ARGS array.

OVMF_CODE_NORMAL="/usr/share/edk2/x64/OVMF_CODE.4m.fd"
OVMF_CODE_SECBOOT="/usr/share/edk2/x64/OVMF_CODE.secboot.4m.fd"
OVMF_VARS_TEMPLATE="/usr/share/edk2/x64/OVMF_VARS.4m.fd"

build_qemu_args() {
  local profile="$1"
  local secure_boot="$2"
  local ssh_port="$3"
  local test_dir="$4"

  QEMU_ARGS=()

  # ── Machine & CPU ────────────────────────────────────────────────────────────
  QEMU_ARGS+=(-enable-kvm)
  QEMU_ARGS+=(-machine q35,vmport=off)
  QEMU_ARGS+=(-cpu host)
  QEMU_ARGS+=(-smp 4,cores=2,threads=2,sockets=1)
  QEMU_ARGS+=(-m 4096)
  QEMU_ARGS+=(-uuid "$(cat /proc/sys/kernel/random/uuid 2>/dev/null || uuidgen)")

  # ── Firmware ─────────────────────────────────────────────────────────────────
  if [[ "$secure_boot" == "on" ]]; then
    local vars_copy
    vars_copy="$(mktemp /tmp/OVMF_VARS_XXXXXXXX.fd)"
    cp "$OVMF_VARS_TEMPLATE" "$vars_copy"
    QEMU_ARGS+=(-machine q35,vmport=off,smm=on)  # override previous -machine
    QEMU_ARGS+=(-global driver=cfi.pflash01,property=secure,value=on)
    QEMU_ARGS+=(-drive "if=pflash,readonly=on,format=raw,file=${OVMF_CODE_SECBOOT}")
    QEMU_ARGS+=(-drive "if=pflash,format=raw,file=${vars_copy}")
    # clean up temp VARS on exit
    trap "rm -f '$vars_copy'" EXIT
  else
    QEMU_ARGS+=(-bios /usr/share/edk2/x64/OVMF.4m.fd)
  fi

  # ── SMBIOS ───────────────────────────────────────────────────────────────────
  # Type 0: BIOS
  QEMU_ARGS+=(-smbios "type=0,vendor=American Megatrends Inc.,version=F14,date=12/01/2021,uefi=on")

  case "$profile" in
    desktop)
      # Type 1: System
      QEMU_ARGS+=(-smbios "type=1,manufacturer=Gigabyte Technology Co. Ltd.,product=B550 AORUS ELITE V2,version=1.0,serial=SN-DESKTOP-001,uuid=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid),sku=Default string,family=Desktop")
      # Type 2: Board
      QEMU_ARGS+=(-smbios "type=2,manufacturer=Gigabyte Technology Co. Ltd.,product=B550 AORUS ELITE V2,version=x.x,serial=SN-BOARD-DESKTOP,asset=Base Board Asset Tag,location=Default string")
      # Type 3: Chassis — desktop (type 3)
      QEMU_ARGS+=(-smbios "type=3,manufacturer=Default string,version=Default string,serial=Default string,asset=Default string,sku=Default string")
      ;;
    laptop)
      # Type 1: System — ThinkPad X1 Carbon Gen 10
      QEMU_ARGS+=(-smbios "type=1,manufacturer=LENOVO,product=21CB,version=ThinkPad X1 Carbon Gen 10,serial=SN-LAPTOP-001,uuid=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid),sku=LENOVO_MT_21CB,family=ThinkPad X1 Carbon Gen 10")
      # Type 2: Board
      QEMU_ARGS+=(-smbios "type=2,manufacturer=LENOVO,product=21CB,version=SDK0J40697 WIN,serial=SN-BOARD-LAPTOP,asset=Not Available,location=Not Available")
      # Type 3: Chassis — notebook (type 10)
      QEMU_ARGS+=(-smbios "type=3,manufacturer=LENOVO,version=None,serial=SN-CHASSIS-LAPTOP,asset=Not Available,sku=Not Available")
      ;;
    *)
      echo "Unknown profile: $profile (expected desktop or laptop)" >&2
      return 1
      ;;
  esac

  # ── Storage — NVMe ───────────────────────────────────────────────────────────
  local nvme_serial
  case "$profile" in
    desktop) nvme_serial="TESTDESKTOP0001" ;;
    laptop)  nvme_serial="TESTLAPTOP00001" ;;
  esac
  QEMU_ARGS+=(-drive "if=none,id=nvme0,format=qcow2,file=${test_dir}/arch-test.qcow2")
  QEMU_ARGS+=(-device "nvme,drive=nvme0,serial=${nvme_serial}")

  # ── Network ──────────────────────────────────────────────────────────────────
  QEMU_ARGS+=(-netdev "user,id=net0,hostfwd=tcp::${ssh_port}-:22")
  QEMU_ARGS+=(-device "e1000e,netdev=net0,mac=52:54:00:12:34:56")

  # ── USB Controller + Input ───────────────────────────────────────────────────
  QEMU_ARGS+=(-device qemu-xhci)
  QEMU_ARGS+=(-device usb-kbd)
  QEMU_ARGS+=(-device usb-tablet)

  # ── Audio ─────────────────────────────────────────────────────────────────────
  QEMU_ARGS+=(-audiodev none,id=audio0)
  QEMU_ARGS+=(-device ich9-intel-hda)
  QEMU_ARGS+=(-device hda-duplex,audiodev=audio0)

  # ── GPU / Display ─────────────────────────────────────────────────────────────
  QEMU_ARGS+=(-device "virtio-vga,xres=1920,yres=1080")
  QEMU_ARGS+=(-display "gtk,zoom-to-fit=on")
}
