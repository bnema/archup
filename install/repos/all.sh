#!/bin/bash
# Repository setup phase

stop_log_output  # Stop for interactive prompts

# AUR Support prompt
gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "AUR Support"
echo

if gum confirm "Enable AUR support?" --padding "0 0 1 $PADDING_LEFT"; then
  export ARCHUP_AUR="enabled"

  # Choose AUR helper
  gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "Choose AUR Helper"
  AUR_HELPER=$(gum choose --cursor.foreground 6 --padding "0 0 1 $PADDING_LEFT" "paru" "yay")

  if [ "$AUR_HELPER" = "paru" ]; then
    source "$ARCHUP_INSTALL/repos/paru.sh"
  else
    source "$ARCHUP_INSTALL/repos/yay.sh"
  fi
else
  gum style --foreground 3 --padding "0 0 1 $PADDING_LEFT" "[SKIP] Skipping AUR support"
  echo "AUR: disabled" >> "$ARCHUP_INSTALL_LOG_FILE"
  export ARCHUP_AUR="disabled"
fi

source "$ARCHUP_INSTALL/repos/chaotic.sh"     # Interactive (gum confirm)
start_log_output # Resume (will be stopped by stop_install_log)
