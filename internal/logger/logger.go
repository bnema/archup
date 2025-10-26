package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bnema/archup/internal/system"
)

// Logger wraps slog.Logger and adds dry-run mode support
type Logger struct {
	slog    *slog.Logger
	logPath string
	dryRun  bool
}

// New creates a new Logger with optional file logging and dry-run mode
func New(logPath string, dryRun bool) (*Logger, error) {
	// Create slog handler
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &Logger{
		slog:    slog.New(handler),
		logPath: logPath,
		dryRun:  dryRun,
	}, nil
}

// NewFromSlog creates a Logger from an existing slog.Logger
func NewFromSlog(slogLogger *slog.Logger, logPath string, dryRun bool) *Logger {
	return &Logger{
		slog:    slogLogger,
		logPath: logPath,
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
