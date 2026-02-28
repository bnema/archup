#!/bin/bash
# Post-boot: Configure firewalld for a standard desktop

set -euo pipefail

LOG_FILE="/var/log/archup-first-boot.log"
log() { echo "[firewalld] $*" | tee -a "$LOG_FILE"; }

log "Configuring firewalld..."

systemctl enable --now firewalld.service

# Use the 'public' zone — sensible defaults for a desktop on untrusted networks
firewall-cmd --set-default-zone=public

# Allow mDNS so LAN discovery (Avahi, printers, etc.) works
firewall-cmd --permanent --zone=public --add-service=mdns

# Allow DHCP client (outgoing requests to get an IP)
firewall-cmd --permanent --zone=public --add-service=dhcpv6-client

# Block everything else incoming — no SSH, no exposed ports by default
firewall-cmd --permanent --zone=public --remove-service=ssh 2>/dev/null || true

firewall-cmd --reload

log "Firewall configured (zone: public, mDNS + DHCPv6 client allowed)"
