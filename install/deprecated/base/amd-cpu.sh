#!/bin/bash
# AMD CPU detection and P-State configuration
# Detects Zen generation and recommends appropriate AMD P-State modes

# This script should only be called for AMD CPUs
# Requires: gum, /proc/cpuinfo

detect_zen_generation() {
  local cpu_family
  local cpu_model

  # Get CPU family and model from /proc/cpuinfo
  cpu_family=$(grep -m1 'cpu family' /proc/cpuinfo | awk '{print $4}')
  cpu_model=$(grep -m1 'model' /proc/cpuinfo | awk '{print $3}')

  # Convert to hex for easier comparison (though we'll use decimal)
  local family_dec=$((cpu_family))
  local model_dec=$((cpu_model))

  # Detect Zen generation based on CPU family and model
  # Reference: https://en.wikichip.org/wiki/amd/cpuid

  local zen_gen=""
  local zen_label=""

  case $family_dec in
    23) # Family 17h
      if [ $model_dec -ge 0 ] && [ $model_dec -le 15 ]; then
        zen_gen="1"
        zen_label="Zen 1"
      elif [ $model_dec -ge 16 ] && [ $model_dec -le 31 ]; then
        zen_gen="1+"
        zen_label="Zen+"
      elif [ $model_dec -ge 48 ] && [ $model_dec -le 127 ]; then
        zen_gen="2"
        zen_label="Zen 2"
      else
        zen_gen="2"
        zen_label="Zen 2 (assumed)"
      fi
      ;;
    25) # Family 19h
      if [ $model_dec -ge 0 ] && [ $model_dec -le 15 ]; then
        zen_gen="3"
        zen_label="Zen 3"
      elif [ $model_dec -ge 32 ] && [ $model_dec -le 47 ]; then
        zen_gen="3"
        zen_label="Zen 3"
      elif [ $model_dec -ge 80 ] && [ $model_dec -le 95 ]; then
        zen_gen="3+"
        zen_label="Zen 3+"
      elif [ $model_dec -ge 96 ] && [ $model_dec -le 175 ]; then
        zen_gen="4"
        zen_label="Zen 4"
      else
        zen_gen="3"
        zen_label="Zen 3 (assumed)"
      fi
      ;;
    26) # Family 1Ah
      zen_gen="5"
      zen_label="Zen 5"
      ;;
    *)
      zen_gen="unknown"
      zen_label="Unknown"
      ;;
  esac

  echo "${zen_gen}|${zen_label}"
}

get_recommended_pstate_modes() {
  local zen_gen=$1
  local kernel_version

  # Get kernel version
  kernel_version=$(uname -r | cut -d'.' -f1,2)
  local major=$(echo "$kernel_version" | cut -d'.' -f1)
  local minor=$(echo "$kernel_version" | cut -d'.' -f2)

  # Calculate kernel version as single number for comparison
  local kernel_num=$((major * 100 + minor))

  # Determine available modes based on Zen generation and kernel version
  # Reference: https://docs.kernel.org/admin-guide/pm/amd-pstate.html
  # Active: CPPC autonomous (amd_pstate_epp) - kernel 6.3+
  # Passive: CPPC non-autonomous (amd_pstate) - all kernels
  # Guided: CPPC guided autonomous (amd_pstate) - kernel 6.4+

  case $zen_gen in
    "1"|"1+")
      # Zen 1 and Zen+: No CPPC support, use acpi-cpufreq
      echo "none"
      ;;
    "2")
      # Zen 2: Basic CPPC support, may need shared_mem=1 parameter
      # Passive: available, Active: 6.3+
      if [ $kernel_num -ge 603 ]; then
        echo "passive|active"
      else
        echo "passive"
      fi
      ;;
    "3"|"3+"|"4"|"5")
      # Zen 3+: Full CPPC support for all modes
      # Passive: available, Active: 6.3+, Guided: 6.4+
      local modes="passive"
      if [ $kernel_num -ge 603 ]; then
        modes="${modes}|active"
      fi
      if [ $kernel_num -ge 604 ]; then
        modes="${modes}|guided"
      fi
      echo "$modes"
      ;;
    *)
      echo "passive"
      ;;
  esac
}

get_pstate_recommendation() {
  local zen_gen=$1

  # Return the best recommended mode for each Zen generation
  case $zen_gen in
    "1"|"1+")
      echo "none" # Use default acpi-cpufreq
      ;;
    "2")
      echo "passive" # Conservative choice for Zen 2
      ;;
    "3"|"3+")
      echo "active" # Best for Zen 3/3+
      ;;
    "4"|"5")
      echo "active" # Best for Zen 4/5
      ;;
    *)
      echo "passive"
      ;;
  esac
}

get_additional_params() {
  local zen_gen=$1
  local mode=$2

  # Some Zen 2 CPUs need shared_mem parameter
  if [ "$zen_gen" = "2" ] && [ "$mode" = "active" ]; then
    echo "amd_pstate.shared_mem=1"
  else
    echo ""
  fi
}

# Main execution
main() {
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AMD CPU Detection"

  # Detect Zen generation
  local zen_info
  zen_info=$(detect_zen_generation)
  local zen_gen=$(echo "$zen_info" | cut -d'|' -f1)
  local zen_label=$(echo "$zen_info" | cut -d'|' -f2)

  # Get CPU model name for display
  local cpu_model_name
  cpu_model_name=$(grep -m1 'model name' /proc/cpuinfo | cut -d':' -f2 | sed 's/^[ \t]*//')

  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Detected: $zen_label"
  gum style --padding "0 0 0 $PADDING_LEFT" "  CPU: $cpu_model_name"

  # Check if Zen generation supports AMD pstate
  if [ "$zen_gen" = "1" ] || [ "$zen_gen" = "1+" ]; then
    gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[WARN] $zen_label does not support AMD P-State driver"
    gum style --padding "0 0 1 $PADDING_LEFT" "  Using default acpi-cpufreq driver"
    export ARCHUP_AMD_PSTATE=""
    export ARCHUP_AMD_KERNEL_PARAMS=""
    echo "AMD CPU: $zen_label (no P-State support)" >> "$ARCHUP_INSTALL_LOG_FILE"
    return 0
  fi

  # Get available modes for this Zen generation
  local available_modes
  available_modes=$(get_recommended_pstate_modes "$zen_gen")

  echo
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AMD P-State Configuration"
  gum style --padding "0 0 0 $PADDING_LEFT" "Select AMD P-State driver mode for CPU frequency scaling"
  gum style --foreground 8 --padding "0 0 0 $PADDING_LEFT" "Note: Requires CPPC enabled in UEFI (AMD CBS > NBIO > SMU > CPPC)"
  echo

  # Build choice menu based on available modes
  declare -a MODE_CHOICES

  if echo "$available_modes" | grep -q "active"; then
    MODE_CHOICES+=("active - CPPC autonomous with EPP hints (best performance, desktop/gaming)")
  fi

  if echo "$available_modes" | grep -q "guided"; then
    MODE_CHOICES+=("guided - CPPC guided autonomous (balanced, laptop/hybrid)")
  fi

  if echo "$available_modes" | grep -q "passive"; then
    MODE_CHOICES+=("passive - CPPC non-autonomous (compatibility, manual governor control)")
  fi

  # Show recommendation
  local recommended
  recommended=$(get_pstate_recommendation "$zen_gen")

  if [ "$recommended" != "none" ]; then
    gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "Recommended for $zen_label: $recommended"
    echo
  fi

  # Let user choose
  local choice
  choice=$(gum choose --padding "0 0 1 $PADDING_LEFT" "${MODE_CHOICES[@]}")

  # Extract mode name
  local selected_mode
  selected_mode=$(echo "$choice" | awk '{print $1}')

  # Get any additional kernel parameters needed
  local extra_params
  extra_params=$(get_additional_params "$zen_gen" "$selected_mode")

  # Export configuration
  export ARCHUP_AMD_PSTATE="$selected_mode"

  if [ -n "$extra_params" ]; then
    export ARCHUP_AMD_KERNEL_PARAMS="amd_pstate=$selected_mode $extra_params"
  else
    export ARCHUP_AMD_KERNEL_PARAMS="amd_pstate=$selected_mode"
  fi

  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] P-State mode: $selected_mode"

  if [ -n "$extra_params" ]; then
    gum style --padding "0 0 1 $PADDING_LEFT" "  Additional params: $extra_params"
  fi

  echo "AMD CPU: $zen_label, P-State: $selected_mode" >> "$ARCHUP_INSTALL_LOG_FILE"

  if [ -n "$extra_params" ]; then
    echo "  Extra params: $extra_params" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
}

# Run if executed directly (not sourced)
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
  main
fi
