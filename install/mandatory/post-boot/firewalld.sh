#!/bin/bash
# Post-boot: Configure firewalld with basic protection

# Enable and start firewalld
systemctl enable --now firewalld.service

# Set default zone to drop (deny all incoming by default)
firewall-cmd --set-default-zone=drop

# Allow outgoing connections (drop zone blocks incoming only)
# Add common services if needed
# firewall-cmd --permanent --add-service=ssh

# Reload to apply changes
firewall-cmd --reload

echo "Firewall configured (default deny incoming)"
