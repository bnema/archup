package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// SlogAdapter implements the Logger port using Go's standard log/slog
type SlogAdapter struct {
	logger  *slog.Logger
	logPath string
}

// NewSlogAdapter creates a new slog-based logger
// If logPath is empty, logs are written to stderr only
func NewSlogAdapter(logPath string) (*SlogAdapter, error) {
	var logger *slog.Logger

	if logPath != "" {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file for appending
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		// Create slog handler with both stderr and file output
		handler := slog.NewTextHandler(file, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger = slog.New(handler)
	} else {
		// Log to stderr only
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger = slog.New(handler)
	}

	return &SlogAdapter{
		logger:  logger,
		logPath: logPath,
	}, nil
}

// Info logs an informational message
func (sa *SlogAdapter) Info(msg string, keysAndValues ...interface{}) {
	sa.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message
func (sa *SlogAdapter) Warn(msg string, keysAndValues ...interface{}) {
	sa.logger.Warn(msg, keysAndValues...)
}

// Error logs an error message
func (sa *SlogAdapter) Error(msg string, keysAndValues ...interface{}) {
	sa.logger.Error(msg, keysAndValues...)
}

// Debug logs a debug message
func (sa *SlogAdapter) Debug(msg string, keysAndValues ...interface{}) {
	sa.logger.Debug(msg, keysAndValues...)
}

// LogPath returns the path to the log file
func (sa *SlogAdapter) LogPath() string {
	return sa.logPath
}
