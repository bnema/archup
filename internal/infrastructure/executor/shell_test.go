package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestShellExecutor_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	output, err := executor.Execute(ctx, "echo", "hello world")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "hello world\n"
	if string(output) != expected {
		t.Errorf("expected %q, got %q", expected, string(output))
	}
}

func TestShellExecutor_Execute_WithArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	// Test with multiple arguments
	output, err := executor.Execute(ctx, "echo", "arg1", "arg2", "arg3")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(string(output), "arg1") {
		t.Errorf("expected arg1 in output, got %s", string(output))
	}
}

func TestShellExecutor_Execute_CommandNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	_, err := executor.Execute(ctx, "nonexistentcommand12345", "arg")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}
}

func TestShellExecutor_Execute_CommandFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	// false command always exits with code 1
	_, err := executor.Execute(ctx, "false")
	if err == nil {
		t.Fatal("expected error for failed command")
	}
}

func TestShellExecutor_ExecuteWithEnv_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	env := map[string]string{
		"TEST_VAR": "test_value",
	}

	// Echo the environment variable
	output, err := executor.ExecuteWithEnv(ctx, env, "sh", "-c", "echo $TEST_VAR")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(string(output), "test_value") {
		t.Errorf("expected test_value in output, got %s", string(output))
	}
}

func TestShellExecutor_ExecuteWithEnv_MultipleVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	output, err := executor.ExecuteWithEnv(ctx, env, "sh", "-c", "echo $VAR1 $VAR2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output_str := string(output)
	if !strings.Contains(output_str, "value1") || !strings.Contains(output_str, "value2") {
		t.Errorf("expected both values in output, got %s", output_str)
	}
}

func TestShellExecutor_ExecuteWithStdin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	output, err := executor.ExecuteWithStdin(ctx, "hello stdin", "cat")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(output) != "hello stdin" {
		t.Errorf("expected stdin to be forwarded, got %q", string(output))
	}
}

func TestShellExecutor_Execute_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// sleep command will be interrupted
	_, err := executor.Execute(ctx, "sleep", "10")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestShellExecutor_Execute_EmptyOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	// true command produces no output
	output, err := executor.Execute(ctx, "true")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(output) != 0 {
		t.Errorf("expected empty output, got %s", string(output))
	}
}

func TestShellExecutor_Execute_StderrCapture(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewShellExecutor(mockLogger)
	ctx := context.Background()

	// sh command that writes to stderr
	_, err := executor.Execute(ctx, "sh", "-c", "echo error >&2; exit 1")
	if err == nil {
		t.Fatal("expected error")
	}

	// Verify stderr is captured in error message
	if !strings.Contains(err.Error(), "error") {
		t.Errorf("expected 'error' in error message, got %s", err.Error())
	}
}
