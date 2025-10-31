package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ScriptExecutor implements the ScriptExecutor port for shell scripts
type ScriptExecutor struct{}

// NewScriptExecutor creates a new script executor
func NewScriptExecutor() *ScriptExecutor {
	return &ScriptExecutor{}
}

// ExecuteScript runs a shell script with environment variables
func (se *ScriptExecutor) ExecuteScript(ctx context.Context, scriptPath string, env map[string]string) error {
	// Verify script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("script not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, "bash", scriptPath)
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
		return fmt.Errorf("script execution failed: %w (stdout: %s, stderr: %s)", err, stdout.String(), stderr.String())
	}

	return nil
}
