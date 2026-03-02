#!/bin/bash
# ArchUp First Boot - Post-boot tasks that require running system

LOG_FILE="/var/log/archup-first-boot.log"

# --- DNS readiness gate ---
wait_for_dns() {
  local host="archlinux.org"
  local timeout=120
  local elapsed=0
  while [ "$elapsed" -le "$timeout" ]; do
    if getent ahosts "$host" >/dev/null 2>&1; then
      echo "[first-boot] DNS ready after ${elapsed}s" | tee -a "$LOG_FILE"
      return 0
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done
  echo "[first-boot] FATAL: DNS not ready after ${timeout}s — aborting" | tee -a "$LOG_FILE"
  return 1
}

wait_for_dns || exit 1

# Display ArchUp logo
if [ -f /usr/local/share/archup/logo.txt ]; then
  cat /usr/local/share/archup/logo.txt
  echo ""
fi

echo "=== ArchUp First Boot Setup - $(date) ===" >> "$LOG_FILE"

# Run snapper configuration
if [ -f /usr/local/share/archup/post-boot/snapper.sh ]; then
  echo "Configuring snapper..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/snapper.sh >> "$LOG_FILE" 2>&1; then
    echo "Snapper configured successfully" >> "$LOG_FILE"
  else
    echo "Snapper configuration failed (non-critical)" >> "$LOG_FILE"
  fi
fi

# Configure firewalld
if [ -f /usr/local/share/archup/post-boot/firewalld.sh ]; then
  echo "Configuring firewall..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/firewalld.sh >> "$LOG_FILE" 2>&1; then
    echo "Firewall configured successfully" >> "$LOG_FILE"
  else
    echo "Firewall configuration failed (non-critical)" >> "$LOG_FILE"
  fi
fi

# Run SSH key generation
if [ -f /usr/local/share/archup/post-boot/ssh-keygen.sh ]; then
  echo "Generating SSH keys..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/ssh-keygen.sh >> "$LOG_FILE" 2>&1; then
    echo "[OK] SSH keys generated successfully" >> "$LOG_FILE"
  else
    echo "[KO] SSH key generation failed (non-critical)" >> "$LOG_FILE"
  fi
fi

# Install archup-cli from GitHub
if [ -f /usr/local/share/archup/post-boot/archup-cli.sh ]; then
  echo "Installing archup-cli from GitHub..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/archup-cli.sh >> "$LOG_FILE" 2>&1; then
    echo "[OK] archup-cli installed successfully" >> "$LOG_FILE"
  else
    echo "[KO] archup-cli installation failed (non-critical)" >> "$LOG_FILE"
  fi
fi

# Install ble.sh (Bash Line Editor)
if [ -f /usr/local/share/archup/post-boot/blesh.sh ]; then
  echo "Installing ble.sh (Bash Line Editor)..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/blesh.sh >> "$LOG_FILE" 2>&1; then
    echo "[OK] ble.sh installed successfully" >> "$LOG_FILE"
  else
    echo "[KO] ble.sh installation failed (non-critical)" >> "$LOG_FILE"
  fi
fi

# Install modern CLI toolkit
if [ -f /usr/local/share/archup/post-boot/cli-tools.sh ]; then
  echo "Installing modern CLI toolkit..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/cli-tools.sh >> "$LOG_FILE" 2>&1; then
    echo "[OK] CLI tools installed" >> "$LOG_FILE"
  else
    echo "[KO] CLI tools installation failed (non-critical)" >> "$LOG_FILE"
  fi
fi

echo "=== First Boot Setup Complete ===" >> "$LOG_FILE"

# Disable this service after first run
systemctl disable archup-first-boot.service >> "$LOG_FILE" 2>&1
