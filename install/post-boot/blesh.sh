#!/bin/bash
# Install and configure ble.sh (Bash Line Editor)
# This provides advanced command-line editing features

set -e

# Get the username from the system
USERNAME=$(ls /home | head -1)

if [[ -z "${USERNAME}" ]]; then
  echo "ERROR: Could not find user for ble.sh installation"
  exit 1
fi

USER_HOME="/home/$USERNAME"
BLESH_INSTALL_DIR="$USER_HOME/.local"

echo "Installing ble.sh for user: $USERNAME"

# Clone ble.sh repository
echo "Cloning ble.sh repository..."
cd /tmp
if [ -d "ble.sh" ]; then
  rm -rf ble.sh
fi

# Run as user to avoid permission issues
sudo -u "$USERNAME" bash << EOF
set -e
cd /tmp
git clone --recursive --depth 1 --shallow-submodules https://github.com/akinomyoga/ble.sh.git

# Build and install ble.sh
echo "Building and installing ble.sh..."
make -C ble.sh install PREFIX="$BLESH_INSTALL_DIR"

# Add ble.sh to .bashrc if not already present
if ! grep -q "blesh/ble.sh" "$USER_HOME/.bashrc"; then
  echo "" >> "$USER_HOME/.bashrc"
  echo "# ble.sh - Bash Line Editor" >> "$USER_HOME/.bashrc"
  echo "[[ \$- == *i* ]] && source -- ~/.local/share/blesh/ble.sh" >> "$USER_HOME/.bashrc"
  echo "ble.sh added to .bashrc"
else
  echo "ble.sh already configured in .bashrc"
fi
EOF

# Cleanup
rm -rf /tmp/ble.sh

echo "[OK] ble.sh installed successfully"
echo "  Location: $BLESH_INSTALL_DIR/share/blesh"
echo "  Note: ble.sh will be active on next shell session"
