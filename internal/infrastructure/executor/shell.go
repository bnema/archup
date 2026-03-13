package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/bnema/archup/internal/domain/ports"
)

// ShellExecutor implements the CommandExecutor port using system shell commands
type ShellExecutor struct {
	logger ports.Logger
}

// NewShellExecutor creates a new shell executor with a logger
func NewShellExecutor(logger ports.Logger) *ShellExecutor {
	return &ShellExecutor{
		logger: logger,
	}
}

// Execute runs a command and returns output
func (se *ShellExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
	return se.run(ctx, "", nil, command, args...)
}

// ExecuteWithStdin runs a command with stdin content and returns output
func (se *ShellExecutor) ExecuteWithStdin(ctx context.Context, stdin string, command string, args ...string) ([]byte, error) {
	return se.run(ctx, stdin, nil, command, args...)
}

// ExecuteWithEnv runs a command with custom environment variables
func (se *ShellExecutor) ExecuteWithEnv(ctx context.Context, env map[string]string, command string, args ...string) ([]byte, error) {
	return se.run(ctx, "", env, command, args...)
}

func (se *ShellExecutor) run(ctx context.Context, stdin string, env map[string]string, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != "" {
		cmd.Stdin = bytes.NewBufferString(stdin)
	}

	if env != nil {
		// Build environment: start with current environment, then add/override with provided env
		cmdEnv := os.Environ()
		for key, value := range env {
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = cmdEnv
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
