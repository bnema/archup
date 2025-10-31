#!/bin/bash
# User Identity Configuration - All identity prompts consolidated

gum style --foreground 6 --padding "1 0 0 $PADDING_LEFT" "User Identity"
gum style --padding "0 0 0 $PADDING_LEFT" "Configure your user account and system details"
echo

# Username
ARCHUP_USERNAME=$(gum input --placeholder "username" \
  --prompt "Username: " \
  --padding "0 0 0 $PADDING_LEFT")

if [ -z "$ARCHUP_USERNAME" ]; then
  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "ERROR: Username cannot be empty"
  exit 1
fi

# Password (with confirmation)
while true; do
  ARCHUP_PASSWORD=$(gum input --password --placeholder "Enter password" \
    --prompt "Password: " \
    --padding "0 0 0 $PADDING_LEFT")

  ARCHUP_PASSWORD_CONFIRM=$(gum input --password --placeholder "Confirm password" \
    --prompt "Confirm: " \
    --padding "0 0 0 $PADDING_LEFT")

  if [ "$ARCHUP_PASSWORD" = "$ARCHUP_PASSWORD_CONFIRM" ]; then
    break
  else
    gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "Passwords do not match. Try again."
  fi
done

# Encryption Password Choice
echo
gum style --foreground 6 --padding "0 0 0 $PADDING_LEFT" "Disk Encryption"
gum style --padding "0 0 0 $PADDING_LEFT" "Choose whether to use the same password for disk encryption"
if gum confirm "Use same password for disk encryption?" --padding "0 0 1 $PADDING_LEFT"; then
  ARCHUP_ENCRYPTION_USE_SAME_PASSWORD="true"
else
  ARCHUP_ENCRYPTION_USE_SAME_PASSWORD="false"
fi

# Email (optional)
echo
gum style --padding "0 0 0 $PADDING_LEFT" "Email is used for SSH key generation and git configuration"
ARCHUP_EMAIL=$(gum input --placeholder "user@example.com (optional)" \
  --prompt "Email: " \
  --padding "0 0 0 $PADDING_LEFT")

# Hostname
echo
ARCHUP_HOSTNAME=$(gum input --placeholder "archup" \
  --prompt "Hostname: " \
  --padding "0 0 0 $PADDING_LEFT" \
  --value "archup")

if [ -z "$ARCHUP_HOSTNAME" ]; then
  ARCHUP_HOSTNAME="archup"
fi

# Timezone (detect via API)
echo
gum style --foreground 6 --padding "0 0 0 $PADDING_LEFT" "Detecting timezone..."
DETECTED_TIMEZONE=$(curl -s "https://ipapi.co/timezone/" 2>/dev/null)

if [ -n "$DETECTED_TIMEZONE" ]; then
  gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "Detected: $DETECTED_TIMEZONE"
  if gum confirm "Use detected timezone?" --padding "0 0 0 $PADDING_LEFT"; then
    ARCHUP_TIMEZONE="$DETECTED_TIMEZONE"
  else
    ARCHUP_TIMEZONE=$(gum input --placeholder "America/New_York" \
      --prompt "Timezone: " \
      --padding "0 0 0 $PADDING_LEFT")
  fi
else
  gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "Unable to detect timezone"
  ARCHUP_TIMEZONE=$(gum input --placeholder "America/New_York" \
    --prompt "Timezone: " \
    --padding "0 0 0 $PADDING_LEFT")
fi

if [ -z "$ARCHUP_TIMEZONE" ]; then
  ARCHUP_TIMEZONE="UTC"
fi

# Export and save all identity variables
export ARCHUP_USERNAME
export ARCHUP_PASSWORD
export ARCHUP_EMAIL
export ARCHUP_HOSTNAME
export ARCHUP_TIMEZONE
export ARCHUP_ENCRYPTION_USE_SAME_PASSWORD

config_set "ARCHUP_USERNAME" "$ARCHUP_USERNAME"
config_set "ARCHUP_PASSWORD" "$ARCHUP_PASSWORD"
config_set "ARCHUP_EMAIL" "$ARCHUP_EMAIL"
config_set "ARCHUP_HOSTNAME" "$ARCHUP_HOSTNAME"
config_set "ARCHUP_TIMEZONE" "$ARCHUP_TIMEZONE"
config_set "ARCHUP_ENCRYPTION_USE_SAME_PASSWORD" "$ARCHUP_ENCRYPTION_USE_SAME_PASSWORD"

# Display summary
echo
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] User: $ARCHUP_USERNAME"
if [ -n "$ARCHUP_EMAIL" ]; then
  gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Email: $ARCHUP_EMAIL"
else
  gum style --foreground 3 --padding "0 0 0 $PADDING_LEFT" "[SKIP] Email: not provided"
fi
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Hostname: $ARCHUP_HOSTNAME"
gum style --foreground 2 --padding "0 0 0 $PADDING_LEFT" "[OK] Timezone: $ARCHUP_TIMEZONE"
if [ "$ARCHUP_ENCRYPTION_USE_SAME_PASSWORD" = "true" ]; then
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Disk Encryption: using account password"
else
  gum style --foreground 2 --padding "0 0 1 $PADDING_LEFT" "[OK] Disk Encryption: separate password (will be prompted during setup)"
fi

echo "User: $ARCHUP_USERNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Email: ${ARCHUP_EMAIL:-<not provided>}" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Hostname: $ARCHUP_HOSTNAME" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Timezone: $ARCHUP_TIMEZONE" >> "$ARCHUP_INSTALL_LOG_FILE"
echo "Encryption Password: $ARCHUP_ENCRYPTION_USE_SAME_PASSWORD" >> "$ARCHUP_INSTALL_LOG_FILE"
