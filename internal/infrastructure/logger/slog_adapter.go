package logger

import (
	"log/slog"
)

// SlogAdapter implements the Logger port using Go's standard log/slog
type SlogAdapter struct {
	logger  *slog.Logger
	logPath string
}

// NewSlogAdapter wraps an existing slog.Logger to implement the Logger port
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{
		logger:  logger,
		logPath: "",
	}
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
