package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/bnema/archup/internal/system"
)

// Logger wraps slog.Logger and adds dry-run mode support
type Logger struct {
	slog     *slog.Logger
	logPath  string
	dryRun   bool
	logFile  *os.File
}

// syncWriter wraps an io.Writer to sync after every write
type syncWriter struct {
	w interface {
		io.Writer
		Sync() error
	}
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	n, err = sw.w.Write(p)
	if err != nil {
		return n, err
	}
	// Sync to disk immediately to ensure logs are written even on crash
	if syncErr := sw.w.Sync(); syncErr != nil {
		// Log sync error but don't fail the write
		fmt.Fprintf(os.Stderr, "Warning: failed to sync log file: %v\n", syncErr)
	}
	return n, err
}

// New creates a new Logger with optional file logging and dry-run mode
func New(logPath string, dryRun bool) (*Logger, error) {
	// Open log file for writing with sync flag for immediate writes
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Wrap log file with sync writer to ensure immediate disk writes
	syncLogFile := &syncWriter{w: logFile}

	// Only write to log file (not stdout) to avoid interfering with TUI
	// If you need stdout logging for debugging, add it back temporarily
	handler := slog.NewTextHandler(syncLogFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		slog:    slog.New(handler),
		logPath: logPath,
		logFile: logFile,
		dryRun:  dryRun,
	}

	// Write initial log entry to verify logging works
	logger.Info("Logger initialized", "log_path", logPath, "dry_run", dryRun)

	return logger, nil
}

// NewFromSlog creates a Logger from an existing slog.Logger
// Note: This does not create a log file - the provided slog.Logger should handle output
func NewFromSlog(slogLogger *slog.Logger, logPath string, dryRun bool) *Logger {
	return &Logger{
		slog:    slogLogger,
		logPath: logPath,
		logFile: nil, // No log file for this constructor
		dryRun:  dryRun,
	}
}

// Info logs an informational message
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// IsDryRun returns true if in dry-run mode
func (l *Logger) IsDryRun() bool {
	return l.dryRun
}

// LogPath returns the log file path
func (l *Logger) LogPath() string {
	return l.logPath
}

// ExecCommand executes a command, or logs it in dry-run mode
func (l *Logger) ExecCommand(command string, args ...string) system.CommandResult {
	if l.dryRun {
		l.slog.Info("DRY-RUN", "command", command, "args", args)
		return system.CommandResult{
			ExitCode: 0,
			Output:   fmt.Sprintf("DRY-RUN: %s %s", command, strings.Join(args, " ")),
			Error:    nil,
		}
	}

	return system.Run(system.RunConfig{
		Command:     command,
		Args:        args,
		StreamToLog: true,
		LogPath:     l.logPath,
	})
}

// ExecSimple executes a command without logging to file
func (l *Logger) ExecSimple(command string, args ...string) system.CommandResult {
	if l.dryRun {
		l.slog.Info("DRY-RUN", "command", command, "args", args)
		return system.CommandResult{
			ExitCode: 0,
			Output:   fmt.Sprintf("DRY-RUN: %s %s", command, strings.Join(args, " ")),
			Error:    nil,
		}
	}

	return system.RunSimple(command, args...)
}

// ExecWithConfig executes a command with custom RunConfig
func (l *Logger) ExecWithConfig(cfg system.RunConfig) system.CommandResult {
	if l.dryRun {
		l.slog.Info("DRY-RUN", "command", cfg.Command, "args", cfg.Args)
		return system.CommandResult{
			ExitCode: 0,
			Output:   fmt.Sprintf("DRY-RUN: %s %s", cfg.Command, strings.Join(cfg.Args, " ")),
			Error:    nil,
		}
	}

	// Set log path if not already set
	if cfg.LogPath == "" {
		cfg.LogPath = l.logPath
		cfg.StreamToLog = true
	}

	return system.Run(cfg)
}

// Slog returns the underlying slog.Logger for advanced usage
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}

// Close closes the log file (call this when done)
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
