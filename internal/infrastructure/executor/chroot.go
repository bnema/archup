package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bnema/archup/internal/domain/ports"
)

// ChrootExecutor implements the ChrootExecutor port using chroot commands
type ChrootExecutor struct {
	logger ports.Logger
}

// NewChrootExecutor creates a new chroot executor with a logger
func NewChrootExecutor(logger ports.Logger) *ChrootExecutor {
	return &ChrootExecutor{
		logger: logger,
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

// ExecuteInChrootWithStdin runs a command inside a chroot with stdin
func (ce *ChrootExecutor) ExecuteInChrootWithStdin(ctx context.Context, chrootPath string, stdin string, command string, args ...string) error {
	if _, err := os.Stat(chrootPath); err != nil {
		return fmt.Errorf("chroot path does not exist: %w", err)
	}

	allArgs := append([]string{chrootPath, command}, args...)
	cmd := exec.CommandContext(ctx, "arch-chroot", allArgs...)
	cmd.Stdin = bytes.NewBufferString(stdin)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chroot command failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// ChrootSystemctl runs systemctl commands in chroot
// This is a convenience method that ensures proper logging of systemctl operations
func (ce *ChrootExecutor) ChrootSystemctl(ctx context.Context, logPath string, chrootPath string, args ...string) error {
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
	cmd := exec.CommandContext(ctx, "arch-chroot", allArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("chroot systemctl failed: %v (stderr: %s)", err, stderr.String())

		// Log to file if specified
		if logPath != "" {
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				if _, writeErr := logFile.WriteString(errMsg + "\n"); writeErr != nil {
					_ = logFile.Close()
					return fmt.Errorf("%s (failed to write log: %v)", errMsg, writeErr)
				}
				if closeErr := logFile.Close(); closeErr != nil {
					return fmt.Errorf("%s (failed to close log: %v)", errMsg, closeErr)
				}
			}
		}

		return fmt.Errorf("%s", errMsg)
	}

	// Log success to file if specified
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			if _, writeErr := fmt.Fprintf(logFile, "systemctl %v executed successfully\n", args); writeErr != nil {
				_ = logFile.Close()
				return fmt.Errorf("failed to write log: %w", writeErr)
			}
			if closeErr := logFile.Close(); closeErr != nil {
				return fmt.Errorf("failed to close log: %w", closeErr)
			}
		}
	}

	return nil
}
