package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ChrootExecutor implements the ChrootExecutor port using chroot commands
type ChrootExecutor struct {
	cmdExecutor *ShellExecutor
}

// NewChrootExecutor creates a new chroot executor
func NewChrootExecutor() *ChrootExecutor {
	return &ChrootExecutor{
		cmdExecutor: NewShellExecutor(),
	}
}

// ExecuteInChroot runs a command inside a chroot environment
func (ce *ChrootExecutor) ExecuteInChroot(ctx context.Context, chrootPath string, command string, args ...string) ([]byte, error) {
	// Verify chroot path exists
	if _, err := os.Stat(chrootPath); err != nil {
		return nil, fmt.Errorf("chroot path does not exist: %w", err)
	}

	// Build chroot command: arch-chroot <path> <command> <args...>
	allArgs := append([]string{chrootPath, command}, args...)
	cmd := exec.CommandContext(ctx, "arch-chroot", allArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("chroot command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ChrootSystemctl runs systemctl commands in chroot
// This is a convenience method that ensures proper logging of systemctl operations
func (ce *ChrootExecutor) ChrootSystemctl(logPath string, chrootPath string, args ...string) error {
	// Verify paths exist
	if _, err := os.Stat(chrootPath); err != nil {
		return fmt.Errorf("chroot path does not exist: %w", err)
	}

	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Build chroot command with systemctl
	allArgs := append([]string{chrootPath, "systemctl"}, args...)
	cmd := exec.Command("arch-chroot", allArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("chroot systemctl failed: %v (stderr: %s)", err, stderr.String())

		// Log to file if specified
		if logPath != "" {
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				defer logFile.Close()
				logFile.WriteString(errMsg + "\n")
			}
		}

		return fmt.Errorf("%s", errMsg)
	}

	// Log success to file if specified
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			defer logFile.Close()
			logFile.WriteString(fmt.Sprintf("systemctl %v executed successfully\n", args))
		}
	}

	return nil
}
