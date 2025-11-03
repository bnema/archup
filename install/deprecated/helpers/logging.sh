#!/bin/bash
# Logging system for ArchUp installer
# All output logged to $ARCHUP_INSTALL_LOG_FILE
# Spinners handle display, errors shown automatically via --show-error

# No-op stubs for compatibility (log monitor removed)
start_log_output() {
  true
}

stop_log_output() {
  true
}

start_install_log() {
  touch "$ARCHUP_INSTALL_LOG_FILE"
  chmod 666 "$ARCHUP_INSTALL_LOG_FILE"

  export ARCHUP_START_TIME=$(date '+%Y-%m-%d %H:%M:%S')
  echo "=== ArchUp Installation Started: $ARCHUP_START_TIME ===" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Don't start log monitor here - run_logged will handle it dynamically
}

stop_install_log() {
  stop_log_output
  show_cursor

  if [[ -n ${ARCHUP_INSTALL_LOG_FILE:-} ]]; then
    ARCHUP_END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    echo "" >> "$ARCHUP_INSTALL_LOG_FILE"
    echo "=== ArchUp Installation Completed: $ARCHUP_END_TIME ===" >> "$ARCHUP_INSTALL_LOG_FILE"

    if [ -n "$ARCHUP_START_TIME" ]; then
      ARCHUP_START_EPOCH=$(date -d "$ARCHUP_START_TIME" +%s)
      ARCHUP_END_EPOCH=$(date -d "$ARCHUP_END_TIME" +%s)
      ARCHUP_DURATION=$((ARCHUP_END_EPOCH - ARCHUP_START_EPOCH))

      ARCHUP_MINS=$((ARCHUP_DURATION / 60))
      ARCHUP_SECS=$((ARCHUP_DURATION % 60))

      echo "Duration: ${ARCHUP_MINS}m ${ARCHUP_SECS}s" >> "$ARCHUP_INSTALL_LOG_FILE"
    fi
    echo "=================================" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi
}

run_logged() {
  local script="$1"

  export CURRENT_SCRIPT="$script"

  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting: $script" >> "$ARCHUP_INSTALL_LOG_FILE"

  # Source helpers in subshell before running the script so functions are available
  # Redirect output to log file only
  bash -c "source '$ARCHUP_INSTALL/helpers/all.sh' && source '$script'" >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  local exit_code=$?

  if [ $exit_code -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Completed: $script" >> "$ARCHUP_INSTALL_LOG_FILE"
    unset CURRENT_SCRIPT
  else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Failed: $script (exit code: $exit_code)" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi

  return $exit_code
}
