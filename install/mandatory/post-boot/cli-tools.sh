#!/bin/bash
# cli-tools.sh — Install modern CLI toolkit (barebone, no DE or theme)
set -euo pipefail

LOG_FILE="/var/log/archup-first-boot.log"
USERNAME=$(getent passwd | awk -F: '$3 >= 1000 && $3 < 65534 && $1 != "nobody" {print $1; exit}')
USER_HOME=$(getent passwd "$USERNAME" | cut -d: -f6)

log() { echo "[cli-tools] $*" | tee -a "$LOG_FILE"; }

if [[ -z "$USERNAME" ]]; then
  log "ERROR: Could not detect user in /home"
  exit 1
fi

log "Installing modern CLI toolkit for $USERNAME..."

log "Refreshing package databases..."
pacman -Sy --noconfirm

pacman -S --needed --noconfirm \
  eza bat fd ripgrep fzf zoxide starship \
  btop yazi git-delta gdu procs tealdeer \
  lazygit atuin man-db neovim wget curl git

log "Configuring shell environment for $USERNAME..."

# Remove any previous archup block (idempotent re-run)
if [ -f "$USER_HOME/.bashrc" ]; then
  sed -i '/# BEGIN archup cli tools/,/# END archup cli tools/d' "$USER_HOME/.bashrc"
fi

# Append archup block
cat >> "$USER_HOME/.bashrc" << 'BASHRC'
# BEGIN archup cli tools
# Aliases
alias ls='eza --icons'
alias ll='eza -la --icons --git'
alias la='eza -a --icons'
alias lt='eza --tree --icons'
alias cat='bat'
alias grep='rg'
alias find='fd'
alias top='btop'
alias vim='nvim'
alias vi='nvim'

# Git delta pager
export GIT_PAGER='delta'

# Man pages rendered with bat
export MANPAGER="sh -c 'col -bx | bat -l man -p'"

# fzf key bindings and fuzzy completion
source <(fzf --bash) 2>/dev/null || true

# zoxide (smart cd with frecency)
eval "$(zoxide init bash)"
[[ $- == *i* ]] && alias cd='z'

# atuin (shell history search)
eval "$(atuin init bash)" 2>/dev/null || true

# starship prompt (no custom theme — uses default)
eval "$(starship init bash)"

# ble.sh (Bash Line Editor) — must be sourced last
[[ $- == *i* ]] && source -- ~/.local/share/blesh/ble.sh 2>/dev/null || true
# END archup cli tools
BASHRC

chown "$USERNAME:$USERNAME" "$USER_HOME/.bashrc"

# Update tldr cache
sudo -u "$USERNAME" tldr --update 2>/dev/null || true

# git delta global config
sudo -u "$USERNAME" git config --global core.pager delta
sudo -u "$USERNAME" git config --global interactive.diffFilter "delta --color-only"
sudo -u "$USERNAME" git config --global delta.navigate true
sudo -u "$USERNAME" git config --global delta.side-by-side true
sudo -u "$USERNAME" git config --global merge.conflictstyle diff3
sudo -u "$USERNAME" git config --global diff.colorMoved default

log "CLI toolkit installed successfully."
