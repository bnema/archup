package logger

import (
	"log/slog"
	"os"
	"testing"
)

func TestSlogAdapter_NewSlogAdapter(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}

	if adapter.LogPath() != "" {
		t.Errorf("expected empty log path, got %s", adapter.LogPath())
	}
}

func TestSlogAdapter_Info(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	// Should not panic
	adapter.Info("test message", "key", "value")
}

func TestSlogAdapter_Warn(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	// Should not panic
	adapter.Warn("warning message", "warn_key", "warn_value")
}

func TestSlogAdapter_Error(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	// Should not panic
	adapter.Error("error message", "error_key", "error_value")
}

func TestSlogAdapter_Debug(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	// Should not panic
	adapter.Debug("debug message", "debug_key", "debug_value")
}

func TestSlogAdapter_MultipleLogCalls(t *testing.T) {
	slogLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	adapter := NewSlogAdapter(slogLogger)

	adapter.Info("first message")
	adapter.Warn("second message")
	adapter.Error("third message")
	adapter.Debug("fourth message")

	// All calls should succeed without panicking
}
