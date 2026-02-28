#!/bin/bash
# cli-tools.sh — Install modern CLI toolkit (barebone, no DE or theme)
set -euo pipefail

LOG_FILE="/var/log/archup-first-boot.log"
USERNAME=$(getent passwd | awk -F: '$3 >= 1000 && $3 < 65534 && $1 != "nobody" {print $1; exit}')
USER_HOME=$(getent passwd "$USERNAME" | cut -d: -f6)

log() { echo "[cli-tools] $*" | tee -a "$LOG_FILE"; }

if [[ -z "$USERNAME" ]]; then
  log "ERROR: Could not detect user via getent passwd"
  exit 1
fi

log "Installing modern CLI toolkit for $USERNAME..."

log "Upgrading system and refreshing package databases..."
pacman -Syu --noconfirm

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
alias rg='rg --color=auto'
alias fd='fd --color=auto'
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

# starship prompt
eval "$(starship init bash)"

# ble.sh (Bash Line Editor) — must be sourced last
[[ $- == *i* ]] && source -- ~/.local/share/blesh/ble.sh 2>/dev/null || true
# END archup cli tools
BASHRC

chown "$USERNAME:$USERNAME" "$USER_HOME/.bashrc"

# Deploy starship config
mkdir -p "$USER_HOME/.config"
cat > "$USER_HOME/.config/starship.toml" << 'STARSHIP'
# Arch-inspired blue color scheme
add_newline = true

[username]
show_always = true
style_user = "fg:#87ceeb"
style_root = "fg:#ff6b8a"
format = "[$user]($style) "

[hostname]
disabled = true

[directory]
style = "fg:#e8f4f8"
format = "in [$path]($style) "

[git_branch]
symbol = " "
style = "fg:#6bb6d6"
format = "on [$symbol$branch]($style) "

[character]
success_symbol = '[❯](fg:#6bb6d6)'
error_symbol = '[✗](fg:#ff6b8a)'
STARSHIP
chown -R "$USERNAME:$USERNAME" "$USER_HOME/.config"

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

# Drop a one-shot hook into ~/.bash_profile so dms-opt-in.sh runs on first
# interactive user login (after the system is fully up) and then removes itself.
DMS_SCRIPT="/usr/local/share/archup/post-boot/dms-opt-in.sh"
if [ -f "$DMS_SCRIPT" ]; then
  touch "$USER_HOME/.bash_profile"
  # Remove any stale archup-dms block (idempotent)
  sed -i '/# BEGIN archup-dms first-login/,/# END archup-dms first-login/d' "$USER_HOME/.bash_profile"
  cat >> "$USER_HOME/.bash_profile" << PROFILE
# BEGIN archup-dms first-login
if [ -f "$DMS_SCRIPT" ]; then
  ARCHUP_USERNAME="$USERNAME" bash "$DMS_SCRIPT"
fi
sed -i '/# BEGIN archup-dms first-login/,/# END archup-dms first-login/d' ~/.bash_profile
# END archup-dms first-login
PROFILE
  chown "$USERNAME:$USERNAME" "$USER_HOME/.bash_profile"
  log "Registered dms-opt-in.sh to run on first user login."
fi
