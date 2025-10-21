#!/bin/bash
# Post-install phase: Final touches after system is configured

source "$ARCHUP_INSTALL/post-install/boot-logo.sh"
source "$ARCHUP_INSTALL/post-install/plymouth.sh"
source "$ARCHUP_INSTALL/post-install/snapper.sh"
source "$ARCHUP_INSTALL/post-install/ufw.sh"
source "$ARCHUP_INSTALL/post-install/pacman.sh"
source "$ARCHUP_INSTALL/post-install/post-boot-setup.sh"
source "$ARCHUP_INSTALL/post-install/hooks.sh"
source "$ARCHUP_INSTALL/post-install/shell-config.sh"
source "$ARCHUP_INSTALL/post-install/verify.sh"
# source "$ARCHUP_INSTALL/post-install/unmount.sh"  # DISABLED FOR DEBUGGING
