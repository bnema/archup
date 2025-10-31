package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScriptExecutor_NewScriptExecutor(t *testing.T) {
	executor := NewScriptExecutor()
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
}

func TestScriptExecutor_ExecuteScript_Success(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a simple test script
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := "#!/bin/bash\nexit 0"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx := context.Background()
	err := executor.ExecuteScript(ctx, scriptPath, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScriptExecutor_ExecuteScript_NotFound(t *testing.T) {
	executor := NewScriptExecutor()
	ctx := context.Background()

	err := executor.ExecuteScript(ctx, "/nonexistent/script.sh", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent script")
	}
}

func TestScriptExecutor_ExecuteScript_Failure(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a script that fails
	scriptPath := filepath.Join(tmpDir, "fail.sh")
	scriptContent := "#!/bin/bash\nexit 1"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx := context.Background()
	err := executor.ExecuteScript(ctx, scriptPath, nil)
	if err == nil {
		t.Fatal("expected error for failed script")
	}
}

func TestScriptExecutor_ExecuteScript_WithEnv(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a script that uses environment variables
	scriptPath := filepath.Join(tmpDir, "env.sh")
	scriptContent := "#!/bin/bash\ntest \"$TEST_VAR\" = \"test_value\" || exit 1"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx := context.Background()
	env := map[string]string{
		"TEST_VAR": "test_value",
	}

	err := executor.ExecuteScript(ctx, scriptPath, env)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScriptExecutor_ExecuteScript_WithMultipleEnv(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a script that checks multiple environment variables
	scriptPath := filepath.Join(tmpDir, "multienv.sh")
	scriptContent := "#!/bin/bash\ntest \"$VAR1\" = \"value1\" && test \"$VAR2\" = \"value2\" || exit 1"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx := context.Background()
	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	err := executor.ExecuteScript(ctx, scriptPath, env)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestScriptExecutor_ExecuteScript_Stderr(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a script that writes to stderr
	scriptPath := filepath.Join(tmpDir, "stderr.sh")
	scriptContent := "#!/bin/bash\necho error >&2\nexit 1"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx := context.Background()
	err := executor.ExecuteScript(ctx, scriptPath, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	// Verify stderr is captured in error message
	if err != nil && err.Error() == "" {
		t.Error("expected error message to contain stderr")
	}
}

func TestScriptExecutor_ExecuteScript_ContextCancellation(t *testing.T) {
	executor := NewScriptExecutor()
	tmpDir := t.TempDir()

	// Create a script that sleeps
	scriptPath := filepath.Join(tmpDir, "sleep.sh")
	scriptContent := "#!/bin/bash\nsleep 10"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := executor.ExecuteScript(ctx, scriptPath, nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
