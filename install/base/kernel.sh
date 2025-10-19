#!/bin/bash
# Kernel selection and microcode installation
# Detects CPU vendor and auto-installs appropriate microcode

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Kernel Selection"

# Ask user to choose kernel
echo
gum style --padding "0 0 0 $PADDING_LEFT" "Choose your kernel:"
gum style --padding "0 0 0 $PADDING_LEFT" "  • linux      - Stable mainline kernel (recommended)"
gum style --padding "0 0 0 $PADDING_LEFT" "  • linux-lts  - Long-term support (maximum stability)"
gum style --padding "0 0 0 $PADDING_LEFT" "  • linux-zen  - Performance-optimized (gaming/desktop)"
echo

KERNEL_CHOICE=$(gum choose --padding "0 0 1 $PADDING_LEFT" "linux" "linux-lts" "linux-zen")
export ARCHUP_KERNEL="$KERNEL_CHOICE"

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Selected: $ARCHUP_KERNEL"
echo "Kernel: $ARCHUP_KERNEL" >> "$ARCHUP_INSTALL_LOG_FILE"

# Detect CPU vendor
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Detecting CPU vendor..."

CPU_VENDOR=$(grep -m1 'vendor_id' /proc/cpuinfo | awk '{print $3}')

if [ "$CPU_VENDOR" = "GenuineIntel" ]; then
  MICROCODE="intel-ucode"
  CPU_TYPE="Intel"
elif [ "$CPU_VENDOR" = "AuthenticAMD" ]; then
  MICROCODE="amd-ucode"
  CPU_TYPE="AMD"
else
  MICROCODE=""
  CPU_TYPE="Unknown"
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[WARN] Unknown CPU vendor: $CPU_VENDOR"
fi

if [ -n "$MICROCODE" ]; then
  export ARCHUP_MICROCODE="$MICROCODE"
  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Detected: $CPU_TYPE CPU"
  gum style --padding "0 0 1 $PADDING_LEFT" "  Microcode: $MICROCODE"
  echo "CPU: $CPU_TYPE, Microcode: $MICROCODE" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  export ARCHUP_MICROCODE=""
  echo "CPU: Unknown, no microcode" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# AMD-specific tuning (P-State driver selection)
if [ "$CPU_TYPE" = "AMD" ]; then
  echo
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AMD CPU Tuning"
  gum style --padding "0 0 0 $PADDING_LEFT" "Select AMD P-State driver mode for CPU frequency scaling"
  echo

  # Detect available AMD P-State modes from the live system
  gum style --padding "0 0 0 $PADDING_LEFT" "Detecting available P-State modes..."

  # Check if amd-pstate driver is loaded and what modes are available
  AVAILABLE_MODES=()

  # Try to read available modes from sysfs (if driver is already loaded)
  if [ -f /sys/devices/system/cpu/amd_pstate/status ]; then
    # Driver is loaded, check available modes
    # Typically: active, passive, guided, disable
    mapfile -t AVAILABLE_MODES < <(cat /sys/devices/system/cpu/amd_pstate/status 2>/dev/null | grep -oE '(active|passive|guided)' || echo "")
  fi

  # If we couldn't detect modes, check if the CPU supports CPPC
  # Most modern AMD Ryzen CPUs support all three modes
  if [ ${#AVAILABLE_MODES[@]} -eq 0 ]; then
    # Check if CPPC is supported (indicates active mode support)
    if grep -q "cppc" /proc/cpuinfo 2>/dev/null || [ -d /sys/devices/system/cpu/cpu0/cpufreq/amd_pstate ]; then
      AVAILABLE_MODES=("active" "guided" "passive")
    else
      # Fallback: older AMD CPUs might only support passive
      AVAILABLE_MODES=("passive")
    fi
  fi

  # Remove duplicates and sort
  mapfile -t AVAILABLE_MODES < <(printf '%s\n' "${AVAILABLE_MODES[@]}" | sort -u)

  # Display detected modes
  if [ ${#AVAILABLE_MODES[@]} -gt 0 ]; then
    gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Detected ${#AVAILABLE_MODES[@]} available mode(s): ${AVAILABLE_MODES[*]}"
  else
    gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "[WARN] Could not detect available modes, using defaults"
    AVAILABLE_MODES=("passive")
  fi

  echo
  gum style --padding "0 0 0 $PADDING_LEFT" "AMD P-State modes:"
  # Only show descriptions for available modes
  for mode in "${AVAILABLE_MODES[@]}"; do
    case "$mode" in
      active)
        gum style --padding "0 0 0 $PADDING_LEFT" "  • active  - CPPC (best performance, desktop/gaming)"
        ;;
      guided)
        gum style --padding "0 0 0 $PADDING_LEFT" "  • guided  - Balanced (laptop/hybrid use)"
        ;;
      passive)
        gum style --padding "0 0 0 $PADDING_LEFT" "  • passive - Acpi-cpufreq replacement (compatibility)"
        ;;
    esac
  done
  echo
  gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "Note: CPU governors will be available after reboot"
  gum style --padding "0 0 1 $PADDING_LEFT" "      and depend on the P-State mode selected"
  echo

  AMD_PSTATE=$(gum choose --padding "0 0 1 $PADDING_LEFT" "${AVAILABLE_MODES[@]}")
  export ARCHUP_AMD_PSTATE="$AMD_PSTATE"
  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] P-State: $AMD_PSTATE"
  echo "AMD P-State: $AMD_PSTATE (kernel parameter)" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Prepare kernel parameters for AMD P-State
  export ARCHUP_AMD_KERNEL_PARAMS="amd_pstate=$AMD_PSTATE"
else
  export ARCHUP_AMD_PSTATE=""
  export ARCHUP_AMD_KERNEL_PARAMS=""
fi

gum style --foreground 2 --padding "1 0 1 $PADDING_LEFT" "[OK] Kernel configuration complete"
