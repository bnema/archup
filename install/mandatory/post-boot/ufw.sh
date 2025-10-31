#!/bin/bash
# Post-boot: Configure UFW (Uncomplicated Firewall) with basic protection

# Set default policies
ufw --force default deny incoming
ufw --force default allow outgoing

# Enable UFW
ufw --force enable

# Enable UFW service to start on boot
systemctl enable ufw.service

echo "Firewall configured (default deny incoming)"
