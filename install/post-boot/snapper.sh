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

# Tweak default snapper configs
sed -i 's/^TIMELINE_CREATE="yes"/TIMELINE_CREATE="no"/' /etc/snapper/configs/root
sed -i 's/^TIMELINE_CREATE="yes"/TIMELINE_CREATE="no"/' /etc/snapper/configs/home
sed -i 's/^NUMBER_LIMIT="50"/NUMBER_LIMIT="5"/' /etc/snapper/configs/root
sed -i 's/^NUMBER_LIMIT="50"/NUMBER_LIMIT="5"/' /etc/snapper/configs/home
sed -i 's/^NUMBER_LIMIT_IMPORTANT="10"/NUMBER_LIMIT_IMPORTANT="5"/' /etc/snapper/configs/root
sed -i 's/^NUMBER_LIMIT_IMPORTANT="10"/NUMBER_LIMIT_IMPORTANT="5"/' /etc/snapper/configs/home

echo "Configured snapper limits"
