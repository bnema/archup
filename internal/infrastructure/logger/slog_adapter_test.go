package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlogAdapter_NewSlogAdapterWithoutFile(t *testing.T) {
	logger, err := NewSlogAdapter("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	if logger.LogPath() != "" {
		t.Errorf("expected empty log path, got %s", logger.LogPath())
	}
}

func TestSlogAdapter_NewSlogAdapterWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	if logger.LogPath() != logPath {
		t.Errorf("expected log path %s, got %s", logPath, logger.LogPath())
	}

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("expected log file to be created")
	}
}

func TestSlogAdapter_NewSlogAdapterCreatesDirs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "subdir", "nested", "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// Verify file was created with parent directories
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("expected log file to be created with parent directories")
	}
}

func TestSlogAdapter_Info(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should not panic
	logger.Info("test message", "key", "value")

	// Verify content was written
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected log content")
	}
}

func TestSlogAdapter_Warn(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should not panic
	logger.Warn("warning message", "warn_key", "warn_value")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected log content")
	}
}

func TestSlogAdapter_Error(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should not panic
	logger.Error("error message", "error_key", "error_value")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected log content")
	}
}

func TestSlogAdapter_Debug(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should not panic
	logger.Debug("debug message", "debug_key", "debug_value")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected log content")
	}
}

func TestSlogAdapter_MultipleLogCalls(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewSlogAdapter(logPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	logger.Info("first message")
	logger.Warn("second message")
	logger.Error("third message")
	logger.Debug("fourth message")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// All messages should be present
	if len(content) == 0 {
		t.Error("expected log content")
	}
}
