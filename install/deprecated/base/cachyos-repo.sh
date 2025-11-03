#!/bin/bash
# Configure CachyOS repository if linux-cachyos kernel is selected

# Only run if linux-cachyos is selected
if [[ "$ARCHUP_KERNEL" != "linux-cachyos" ]]; then
  exit 0
fi

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring CachyOS repository..."

# Download and import CachyOS keyring
echo "Downloading CachyOS keyring..." >> "$ARCHUP_INSTALL_LOG_FILE"
pacman-key --recv-keys F3B607488DB35A47 --keyserver keyserver.ubuntu.com >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
pacman-key --lsign-key F3B607488DB35A47 >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Detect CPU architecture for optimal repo selection
CPU_LEVEL="x86-64"
if grep -q "avx2" /proc/cpuinfo && grep -q "fma" /proc/cpuinfo; then
  CPU_LEVEL="x86-64-v3"
  echo "Detected x86-64-v3 CPU support" >> "$ARCHUP_INSTALL_LOG_FILE"
else
  echo "Using x86-64 (generic) repository" >> "$ARCHUP_INSTALL_LOG_FILE"
fi

# Add CachyOS repository to live system's pacman.conf (for pacstrap)
echo "Configuring CachyOS repository on live system..." >> "$ARCHUP_INSTALL_LOG_FILE"

# Check if already configured
if ! grep -q "\[cachyos\]" /etc/pacman.conf; then
  # Add CachyOS repo before [core] repository
  sed -i '/^\[core\]/i \
# CachyOS repositories\
[cachyos]\
Include = /etc/pacman.d/cachyos-mirrorlist\
' /etc/pacman.conf
fi

# Create mirrorlist for CachyOS
cat > /etc/pacman.d/cachyos-mirrorlist << 'EOF'
## CachyOS mirrorlist
Server = https://mirror.cachyos.org/repo/$arch/$repo
EOF

# Sync databases
echo "Syncing package databases..." >> "$ARCHUP_INSTALL_LOG_FILE"
pacman -Sy >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] CachyOS repository configured"
echo "CachyOS repository enabled for $CPU_LEVEL architecture" >> "$ARCHUP_INSTALL_LOG_FILE"
