#!/bin/bash
# ArchUp First Boot - Post-boot tasks that require running system

LOG_FILE="/var/log/archup-first-boot.log"

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

# Configure UFW firewall
if [ -f /usr/local/share/archup/post-boot/ufw.sh ]; then
  echo "Configuring firewall..." >> "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/ufw.sh >> "$LOG_FILE" 2>&1; then
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
  echo "Installing ble.sh (Bash Line Editor)..." | tee -a "$LOG_FILE"
  BLESH_OUTPUT=$(bash /usr/local/share/archup/post-boot/blesh.sh 2>&1 | tee -a "$LOG_FILE")
  if [ $? -eq 0 ]; then
    echo "[OK] ble.sh installed successfully" | tee -a "$LOG_FILE"
    # Check if it was added to bashrc
    if echo "$BLESH_OUTPUT" | grep -q "ble.sh added to .bashrc"; then
      echo "[OK] ble.sh configured in ~/.bashrc" | tee -a "$LOG_FILE"
    elif echo "$BLESH_OUTPUT" | grep -q "ble.sh already configured"; then
      echo "[OK] ble.sh already present in ~/.bashrc" | tee -a "$LOG_FILE"
    fi
  else
    echo "[KO] ble.sh installation failed (non-critical)" | tee -a "$LOG_FILE"
  fi
fi

echo "=== First Boot Setup Complete ===" >> "$LOG_FILE"

# Disable this service after first run
systemctl disable archup-first-boot.service >> "$LOG_FILE" 2>&1
