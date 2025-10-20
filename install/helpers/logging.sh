#!/bin/bash
# Logging system for ArchUp installer - fixes variable export issues
# Uses background process for log display instead of exec redirection

# Global variable for monitor background process
monitor_pid=""

# Background process to monitor and display log in real-time
start_log_output() {
  local ANSI_SAVE_CURSOR="\033[s"
  local ANSI_RESTORE_CURSOR="\033[u"
  local ANSI_CLEAR_BELOW="\033[0J"
  local ANSI_HIDE_CURSOR="\033[?25l"
  local ANSI_RESET="\033[0m"
  local ANSI_GRAY="\033[90m"

  printf $ANSI_SAVE_CURSOR
  printf $ANSI_HIDE_CURSOR

  (
    local log_lines=20
    local max_line_width=$((LOGO_WIDTH - 4))

    while true; do
      mapfile -t current_lines < <(tail -n $log_lines "$ARCHUP_INSTALL_LOG_FILE" 2>/dev/null)

      printf "${ANSI_RESTORE_CURSOR}${ANSI_CLEAR_BELOW}"

      for ((i = 0; i < log_lines; i++)); do
        line="${current_lines[i]:-}"

        if [ ${#line} -gt $max_line_width ]; then
          line="${line:0:$max_line_width}..."
        fi

        if [ -n "$line" ]; then
          printf "${ANSI_GRAY}${PADDING_LEFT_SPACES}  → ${line}${ANSI_RESET}\n"
        else
          printf "${PADDING_LEFT_SPACES}\n"
        fi
      done

      sleep 0.1
    done
  ) &
  monitor_pid=$!
}

stop_log_output() {
  if [ -n "${monitor_pid:-}" ]; then
    kill $monitor_pid 2>/dev/null || true
    wait $monitor_pid 2>/dev/null || true
    unset monitor_pid

    # Clean up terminal state
    printf "\033[?25h"  # Show cursor
  fi
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

  # Don't start/stop log monitor here - it runs continuously once started
  # Use bash -c to create a clean subshell and redirect output to log
  bash -c "source '$script'" </dev/null >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  local exit_code=$?

  if [ $exit_code -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Completed: $script" >> "$ARCHUP_INSTALL_LOG_FILE"
    unset CURRENT_SCRIPT
  else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Failed: $script (exit code: $exit_code)" >> "$ARCHUP_INSTALL_LOG_FILE"
  fi

  return $exit_code
}
