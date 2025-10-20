#!/bin/bash
# Error handling for ArchUp installer - improved with logging and interactivity

# Track error handling state
ERROR_HANDLING=false

# Show cursor
show_cursor() {
  printf "\033[?25h"
}

# Hide cursor
hide_cursor() {
  printf "\033[?25l"
}

# Display truncated log lines
show_log_tail() {
  if [[ -f $ARCHUP_INSTALL_LOG_FILE ]]; then
    local log_lines=$((TERM_HEIGHT - LOGO_HEIGHT - 35))
    local max_line_width=$((LOGO_WIDTH - 4))

    echo
    gum style --foreground 3 "Recent log entries:"
    echo

    tail -n $log_lines "$ARCHUP_INSTALL_LOG_FILE" 2>/dev/null | while IFS= read -r line; do
      if ((${#line} > max_line_width)); then
        local truncated_line="${line:0:$max_line_width}..."
      else
        local truncated_line="$line"
      fi

      gum style "$truncated_line"
    done

    echo
  fi
}

# Display failed script or command
show_failed_script_or_command() {
  if [[ -n ${CURRENT_SCRIPT:-} ]]; then
    gum style "Failed script: $CURRENT_SCRIPT"
  else
    local cmd="$BASH_COMMAND"
    local max_cmd_width=$((LOGO_WIDTH - 4))

    if ((${#cmd} > max_cmd_width)); then
      cmd="${cmd:0:$max_cmd_width}..."
    fi

    gum style "Command: $cmd"
  fi
}

# Save original file descriptors
save_original_outputs() {
  exec 3>&1 4>&2
}

# Restore original file descriptors
restore_outputs() {
  if [ -e /proc/self/fd/3 ] && [ -e /proc/self/fd/4 ]; then
    exec 1>&3 2>&4
  fi
}

# Main error handler
catch_errors() {
  # Prevent recursive error handling
  if [[ $ERROR_HANDLING == true ]]; then
    return
  else
    ERROR_HANDLING=true
  fi

  # Store exit code immediately (BEFORE any other commands!)
  local exit_code=$?
  local failed_command="${BASH_COMMAND}"
  local failed_script="${CURRENT_SCRIPT:-}"

  # Exit code 130 means user pressed Ctrl+C - handle gracefully
  if [[ $exit_code -eq 130 ]]; then
    ERROR_HANDLING=false
    interrupt_handler
    return
  fi

  # Log error to file IMMEDIATELY with direct file write
  # This must happen before any stdout/stderr restoration
  {
    echo ""
    echo "=== INSTALLATION ERROR ==="
    echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "Exit Code: $exit_code"
    if [[ -n "$failed_script" ]]; then
      echo "Failed Script: $failed_script"
    fi
    echo "Failed Command: $failed_command"
    echo "=== END ERROR ==="
    echo ""
  } >> "$ARCHUP_INSTALL_LOG_FILE" 2>&1

  # Stop background log monitor
  stop_log_output
  restore_outputs

  # Clear and show error screen
  clear_logo
  show_cursor

  gum style --foreground 1 --padding "1 0 1 $PADDING_LEFT" "ArchUp installation stopped!"
  show_log_tail

  echo
  gum style "This command halted with exit code $exit_code:"
  show_failed_script_or_command
  echo

  gum style "Check logs: $ARCHUP_INSTALL_LOG_FILE"
  gum style "Report issues: ${ARCHUP_REPO_URL:-https://github.com/bnema/ArchUp}/issues"
  echo

  # Interactive menu
  while true; do
    options=("View full log" "Exit")

    choice=$(gum choose "${options[@]}" --header "What would you like to do?" --height 4 --padding "0 0 0 $PADDING_LEFT")

    case "$choice" in
    "View full log")
      if command -v less &>/dev/null; then
        less "$ARCHUP_INSTALL_LOG_FILE"
      else
        cat "$ARCHUP_INSTALL_LOG_FILE" | tail -n 100
      fi
      clear_logo
      ;;
    "Exit" | "")
      exit 1
      ;;
    esac
  done
}

# Interrupt handler (Ctrl+C)
interrupt_handler() {
  if [[ $ERROR_HANDLING == true ]]; then
    return
  fi

  ERROR_HANDLING=true

  stop_log_output
  restore_outputs
  show_cursor

  echo
  gum style --foreground 3 --padding "1 0 1 $PADDING_LEFT" "Installation interrupted by user"
  echo "Installation cancelled at: $(date '+%Y-%m-%d %H:%M:%S')" >> "$ARCHUP_INSTALL_LOG_FILE"
  echo

  exit 130
}

# Exit handler
exit_handler() {
  local exit_code=$?

  # Only run if exiting with error and not already handled
  if [[ $exit_code -ne 0 && $ERROR_HANDLING != true ]]; then
    catch_errors
  else
    stop_log_output
    show_cursor
  fi
}

# Set up traps
trap catch_errors ERR TERM
trap interrupt_handler INT
trap exit_handler EXIT

# Save original outputs for restoration
save_original_outputs
