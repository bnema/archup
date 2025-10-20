#!/bin/bash
# Configure bash shell with archup default structure

# Get the username created during installation
USERNAME=$(arch-chroot /mnt ls /home | head -1)

if [[ -z "${USERNAME}" ]]; then
  echo "ERROR: Could not find user for shell configuration" >> "$ARCHUP_INSTALL_LOG_FILE"
  exit 1
fi

USER_HOME="/mnt/home/$USERNAME"
ARCHUP_DEFAULT="$USER_HOME/.local/share/archup/default/bash"

echo "Configuring bash shell for $USERNAME..." >> "$ARCHUP_INSTALL_LOG_FILE"

# Create archup default directory structure
mkdir -p "$ARCHUP_DEFAULT"

# Create shell configuration (history, bash-completion, PATH)
cat > "$ARCHUP_DEFAULT/shell" << 'SHELL_END'
# History control
shopt -s histappend
HISTCONTROL=ignoreboth
HISTSIZE=32768
HISTFILESIZE="${HISTSIZE}"

# Bash completion
if [[ ! -v BASH_COMPLETION_VERSINFO && -f /usr/share/bash-completion/bash_completion ]]; then
  source /usr/share/bash-completion/bash_completion
fi

# Set complete path
export PATH="$HOME/.local/bin:$PATH"
set +h
SHELL_END

# Create init configuration (starship, zoxide, fzf)
cat > "$ARCHUP_DEFAULT/init" << 'INIT_END'
# Starship prompt
if command -v starship &> /dev/null; then
  eval "$(starship init bash)"
fi

# Zoxide (smarter cd)
if command -v zoxide &> /dev/null; then
  eval "$(zoxide init bash)"
fi

# FZF fuzzy finder
if command -v fzf &> /dev/null; then
  eval "$(fzf --bash)"
fi
INIT_END

# Create aliases configuration
cat > "$ARCHUP_DEFAULT/aliases" << 'ALIASES_END'
# Eza aliases (modern ls replacement)
if command -v eza &> /dev/null; then
  alias ls='eza -lh --group-directories-first --icons=auto'
  alias lsa='ls -a'
  alias lt='eza --tree --level=2 --long --icons --git'
  alias lta='lt -a'
  alias ll='eza -l --icons=auto'
  alias la='eza -la --icons=auto'
fi

# Bat + FZF integration
if command -v bat &> /dev/null && command -v fzf &> /dev/null; then
  alias ff="fzf --preview 'bat --style=numbers --color=always {}'"
fi

# Bat aliases
if command -v bat &> /dev/null; then
  alias cat='bat --paging=never'
  alias less='bat'
fi

# Directories
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'

# Git shortcuts
alias g='git'
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline --graph --decorate'
ALIASES_END

# Create envs configuration (environment variables)
cat > "$ARCHUP_DEFAULT/envs" << 'ENVS_END'
# Bat configuration
if command -v bat &> /dev/null; then
  export BAT_THEME="TwoDark"
fi

# FZF default options
if command -v fzf &> /dev/null && command -v bat &> /dev/null; then
  export FZF_DEFAULT_OPTS="--preview 'bat --color=always --style=numbers --line-range=:500 {}' --preview-window right:60%:wrap"
fi
ENVS_END

# Create functions configuration
cat > "$ARCHUP_DEFAULT/functions" << 'FUNCTIONS_END'
# Zoxide cd wrapper
if command -v zoxide &> /dev/null; then
  zd() {
    if [ $# -eq 0 ]; then
      builtin cd ~ && return
    elif [ -d "$1" ]; then
      builtin cd "$1"
    else
      z "$@" || echo "Error: Directory not found"
    fi
  }
  alias cd="zd"
fi
FUNCTIONS_END

# Create rc file that sources everything
cat > "$ARCHUP_DEFAULT/rc" << 'RC_END'
source ~/.local/share/archup/default/bash/shell
source ~/.local/share/archup/default/bash/init
source ~/.local/share/archup/default/bash/aliases
source ~/.local/share/archup/default/bash/envs
source ~/.local/share/archup/default/bash/functions
RC_END

# Create user .bashrc that sources archup defaults
cat > "$USER_HOME/.bashrc" << 'BASHRC_END'
# All the default ArchUp aliases and functions
# (don't mess with these directly, just override them here!)
source ~/.local/share/archup/default/bash/rc

# Add your own exports, aliases, and functions here.
#
# Make an alias for invoking commands you use constantly
# alias p='python'
BASHRC_END

# Configure git to use delta for diffs
arch-chroot /mnt su - "$USERNAME" -c "git config --global core.pager delta" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global interactive.diffFilter 'delta --color-only'" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global delta.navigate true" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global delta.light false" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global delta.side-by-side true" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global merge.conflictstyle diff3" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt su - "$USERNAME" -c "git config --global diff.colorMoved default" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

# Set ownership
arch-chroot /mnt chown -R "$USERNAME:$USERNAME" "$USER_HOME/.local" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1
arch-chroot /mnt chown "$USERNAME:$USERNAME" "$USER_HOME/.bashrc" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Shell configured"
echo "Configured shell for user: $USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
