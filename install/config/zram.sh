#!/bin/bash
# Configure zram compressed swap

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Configuring zram swap..."

# Create zram-generator config
cat > /mnt/etc/systemd/zram-generator.conf << 'EOF'
[zram0]
zram-size = min(ram / 2, 4096)
compression-algorithm = zstd
EOF

echo "Created /etc/systemd/zram-generator.conf" >> "$ARCHUP_INSTALL_LOG_FILE"

# Create sysctl config for zram optimization
cat > /mnt/etc/sysctl.d/99-vm-zram-parameters.conf << 'EOF'
vm.swappiness = 180
vm.watermark_boost_factor = 0
vm.watermark_scale_factor = 125
vm.page-cluster = 0
EOF

echo "Created /etc/sysctl.d/99-vm-zram-parameters.conf with optimized settings" >> "$ARCHUP_INSTALL_LOG_FILE"

gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] zram swap configured"
