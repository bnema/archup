#!/bin/bash
# Post-boot: Configure snapper (requires D-Bus)

# Create snapper configs if they don't exist
if ! snapper list-configs 2>/dev/null | grep -q "root"; then
  snapper -c root create-config /
  echo "Created snapper config: root"
fi

if ! snapper list-configs 2>/dev/null | grep -q "home"; then
  snapper -c home create-config /home
  echo "Created snapper config: home"
fi

# Configure timeline snapshots for root: 1 hourly, 1 daily, 1 weekly, 1 monthly, 0 yearly
snapper -c root set-config "TIMELINE_CREATE=yes"
snapper -c root set-config "TIMELINE_LIMIT_HOURLY=1"
snapper -c root set-config "TIMELINE_LIMIT_DAILY=1"
snapper -c root set-config "TIMELINE_LIMIT_WEEKLY=1"
snapper -c root set-config "TIMELINE_LIMIT_MONTHLY=1"
snapper -c root set-config "TIMELINE_LIMIT_YEARLY=0"
snapper -c root set-config "NUMBER_LIMIT=5"
snapper -c root set-config "NUMBER_LIMIT_IMPORTANT=5"

# Disable timeline for home (less critical)
snapper -c home set-config "TIMELINE_CREATE=no"
snapper -c home set-config "NUMBER_LIMIT=5"
snapper -c home set-config "NUMBER_LIMIT_IMPORTANT=5"

# Enable snapper timers
systemctl enable --now snapper-timeline.timer
systemctl enable --now snapper-cleanup.timer

echo "Configured snapper timeline (1h/1d/1w/1m) with snap-pac enabled"
