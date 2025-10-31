//go:generate mockgen -destination=mocks/mock_ports.go -package=mocks . FileSystem,File,CommandExecutor,ChrootExecutor,ScriptExecutor,HTTPClient,Response,Logger,InstallationRepository

package ports

import (
	"context"
	"os"
)

// FileSystem is the port for file system operations
// Implementations must support reading, writing, and file metadata operations
type FileSystem interface {
	// Stat returns file info for the given path
	Stat(name string) (os.FileInfo, error)

	// ReadFile reads entire file content
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to a file with the given permissions
	WriteFile(name string, data []byte, perm os.FileMode) error

	// Create creates or truncates a file
	Create(name string) (File, error)

	// Chmod changes file permissions
	Chmod(name string, perm os.FileMode) error

	// MkdirAll creates directory path with parents
	MkdirAll(path string, perm os.FileMode) error

	// RemoveAll removes a file or directory recursively
	RemoveAll(path string) error

	// Exists checks if a file/directory exists
	Exists(path string) (bool, error)
}

// File is a file handle for reading/writing
type File interface {
	// Read reads up to len(b) bytes from the file
	Read(b []byte) (n int, err error)

	// Write writes len(b) bytes to the file
	Write(b []byte) (n int, err error)

	// Close closes the file
	Close() error
}

// CommandExecutor is the port for executing system commands
type CommandExecutor interface {
	// Execute runs a command and returns output
	Execute(ctx context.Context, command string, args ...string) ([]byte, error)

	// ExecuteWithEnv runs a command with custom environment variables
	ExecuteWithEnv(ctx context.Context, env map[string]string, command string, args ...string) ([]byte, error)
}

// ChrootExecutor is the port for executing commands in a chroot environment
type ChrootExecutor interface {
	// ExecuteInChroot runs a command inside a chroot
	ExecuteInChroot(ctx context.Context, chrootPath string, command string, args ...string) ([]byte, error)

	// ChrootSystemctl runs systemctl commands in chroot
	ChrootSystemctl(logPath string, chrootPath string, args ...string) error
}

// ScriptExecutor is the port for executing shell scripts
type ScriptExecutor interface {
	// ExecuteScript runs a shell script with environment variables
	ExecuteScript(ctx context.Context, scriptPath string, env map[string]string) error
}

// HTTPClient is the port for HTTP operations
type HTTPClient interface {
	// Get performs an HTTP GET request
	Get(url string) (Response, error)

	// Post performs an HTTP POST request
	Post(url string, contentType string, body []byte) (Response, error)
}

// Response represents an HTTP response
type Response interface {
	// StatusCode returns the HTTP status code
	StatusCode() int

	// Body returns the response body
	Body() []byte

	// Close closes the response
	Close() error
}

// Logger is the port for logging
type Logger interface {
	// Info logs an informational message
	Info(msg string, keysAndValues ...any)

	// Warn logs a warning message
	Warn(msg string, keysAndValues ...any)

	// Error logs an error message
	Error(msg string, keysAndValues ...any)

	// Debug logs a debug message
	Debug(msg string, keysAndValues ...any)

	// LogPath returns the path to the log file
	LogPath() string
}

// InstallationRepository is the port for installation persistence
type InstallationRepository interface {
	// Save persists the installation state
	Save(ctx context.Context, installationID string, state string) error

	// Load retrieves the installation state
	Load(ctx context.Context, installationID string) (string, error)

	// Exists checks if an installation exists
	Exists(ctx context.Context, installationID string) (bool, error)
}
