#!/bin/bash
# ArchUp Configuration Manager - .env-based config system
# Stores preflight settings in KEY=VALUE format for reliable access across scripts

# Initialize empty config file if it doesn't exist
init_config() {
  if [ ! -f "$ARCHUP_INSTALL_CONFIG" ]; then
    touch "$ARCHUP_INSTALL_CONFIG"
    chmod 600 "$ARCHUP_INSTALL_CONFIG"
  fi
  # Ensure permissions are restricted (owner read/write only)
  chmod 600 "$ARCHUP_INSTALL_CONFIG" 2>/dev/null || true
}

# Write a value to config
# Usage: config_set "ARCHUP_USERNAME" "brice"
config_set() {
  local key="$1"
  local value="$2"

  init_config

  # Remove existing key if present, then append new value
  sed -i "/^${key}=/d" "$ARCHUP_INSTALL_CONFIG"
  echo "${key}=${value}" >> "$ARCHUP_INSTALL_CONFIG"
}

# Read a value from config
# Usage: config_get "ARCHUP_USERNAME"
config_get() {
  local key="$1"

  init_config

  grep "^${key}=" "$ARCHUP_INSTALL_CONFIG" | cut -d'=' -f2- || echo ""
}

# Source entire config file into current shell
# Usage: config_source
config_source() {
  init_config

  if [ -f "$ARCHUP_INSTALL_CONFIG" ]; then
    # Source the file but handle values with spaces/special chars
    set -a
    source "$ARCHUP_INSTALL_CONFIG"
    set +a
  fi
}

# Display config (for debugging)
config_show() {
  init_config
  if [ -f "$ARCHUP_INSTALL_CONFIG" ]; then
    cat "$ARCHUP_INSTALL_CONFIG"
  fi
}

# Clear config
config_clear() {
  rm -f "$ARCHUP_INSTALL_CONFIG"
  init_config
}
