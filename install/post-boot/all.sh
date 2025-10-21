#!/bin/bash
# ArchUp First Boot - Post-boot tasks that require running system

LOG_FILE="/var/log/archup-first-boot.log"

# Display ArchUp logo
if [ -f /usr/local/share/archup/logo.txt ]; then
  cat /usr/local/share/archup/logo.txt
  echo ""
fi

echo "=== ArchUp First Boot Setup - $(date) ===" | tee "$LOG_FILE"
echo ""

# Run snapper configuration
if [ -f /usr/local/share/archup/post-boot/snapper.sh ]; then
  echo "Configuring snapper..." | tee -a "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/snapper.sh >> "$LOG_FILE" 2>&1; then
    echo "Snapper configured successfully" | tee -a "$LOG_FILE"
  else
    echo "Snapper configuration failed (non-critical)" | tee -a "$LOG_FILE"
  fi
fi

# Configure UFW firewall
if [ -f /usr/local/share/archup/post-boot/ufw.sh ]; then
  echo "Configuring firewall..." | tee -a "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/ufw.sh >> "$LOG_FILE" 2>&1; then
    echo "Firewall configured successfully" | tee -a "$LOG_FILE"
  else
    echo "Firewall configuration failed (non-critical)" | tee -a "$LOG_FILE"
  fi
fi

# Run SSH key generation
if [ -f /usr/local/share/archup/post-boot/ssh-keygen.sh ]; then
  echo "Generating SSH keys..." | tee -a "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/ssh-keygen.sh >> "$LOG_FILE" 2>&1; then
    echo "✓ SSH keys generated successfully" | tee -a "$LOG_FILE"
  else
    echo "✗ SSH key generation failed (non-critical)" | tee -a "$LOG_FILE"
  fi
fi

# Install archup-cli from GitHub
if [ -f /usr/local/share/archup/post-boot/archup-cli.sh ]; then
  echo "Installing archup-cli from GitHub..." | tee -a "$LOG_FILE"
  if bash /usr/local/share/archup/post-boot/archup-cli.sh >> "$LOG_FILE" 2>&1; then
    echo "✓ archup-cli installed successfully" | tee -a "$LOG_FILE"
  else
    echo "✗ archup-cli installation failed (non-critical)" | tee -a "$LOG_FILE"
  fi
fi

echo "=== First Boot Setup Complete ===" >> "$LOG_FILE"

# Disable this service after first run
systemctl disable archup-first-boot.service

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ ArchUp is ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Next: Run 'archup wizard' to configure your system"
echo ""
