#!/bin/bash
# Post-boot: Generate SSH key without password if email was provided

# Read from environment variables (set by systemd service)
EMAIL="$ARCHUP_EMAIL"
USERNAME="$ARCHUP_USERNAME"

if [ -z "$EMAIL" ] || [ -z "$USERNAME" ]; then
  echo "Skipping SSH key generation (no email provided)"
  exit 0
fi

# Get user home directory
USER_HOME=$(eval echo ~"$USERNAME")
SSH_DIR="$USER_HOME/.ssh"

# Create .ssh directory if it doesn't exist
if [ ! -d "$SSH_DIR" ]; then
  mkdir -p "$SSH_DIR"
  chmod 700 "$SSH_DIR"
  chown "$USERNAME:$USERNAME" "$SSH_DIR"
fi

# Generate SSH key if it doesn't exist
if [ ! -f "$SSH_DIR/id_ed25519" ]; then
  echo "Generating SSH key for $USERNAME..."
  su - "$USERNAME" -c "ssh-keygen -t ed25519 -C '$EMAIL' -f '$SSH_DIR/id_ed25519' -N ''"
  echo "SSH key generated at $SSH_DIR/id_ed25519"
else
  echo "SSH key already exists, skipping generation"
fi

# Set proper permissions
chmod 600 "$SSH_DIR/id_ed25519"
chmod 644 "$SSH_DIR/id_ed25519.pub"
chown "$USERNAME:$USERNAME" "$SSH_DIR/id_ed25519" "$SSH_DIR/id_ed25519.pub"

# Configure git with user email and name
echo "Configuring git..."
su - "$USERNAME" -c "git config --global user.email '$EMAIL'"
su - "$USERNAME" -c "git config --global user.name '$USERNAME'"
echo "Git configured with email: $EMAIL and name: $USERNAME"

echo "SSH key setup complete"
