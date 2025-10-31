package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ShellExecutor implements the CommandExecutor port using system shell commands
type ShellExecutor struct{}

// NewShellExecutor creates a new shell executor
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{}
}

// Execute runs a command and returns output
func (se *ShellExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ExecuteWithEnv runs a command with custom environment variables
func (se *ShellExecutor) ExecuteWithEnv(ctx context.Context, env map[string]string, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Build environment: start with current environment, then add/override with provided env
	cmdEnv := os.Environ()
	for key, value := range env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", key, value))
	}
	cmd.Env = cmdEnv

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
