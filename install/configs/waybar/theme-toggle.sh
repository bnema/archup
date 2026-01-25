#!/usr/bin/env bash
set -euo pipefail

current=$(gsettings get org.gnome.desktop.interface color-scheme | tr -d "'")

if [[ "$current" == "prefer-dark" ]]; then
  gsettings set org.gnome.desktop.interface color-scheme default
  gsettings set org.gnome.desktop.interface gtk-theme adw-gtk3
else
  gsettings set org.gnome.desktop.interface color-scheme prefer-dark
  gsettings set org.gnome.desktop.interface gtk-theme adw-gtk3-dark
fi

exit 0
