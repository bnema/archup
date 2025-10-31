package interfaces

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/bnema/archup/internal/system"
)

// FileSystem abstracts file system operations for testing
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
	Open(name string) (io.ReadCloser, error)
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	Create(name string) (io.WriteCloser, error)
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Chmod(name string, mode os.FileMode) error
}

// CommandExecutor abstracts command execution
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
}

// SystemExecutor abstracts system package functions
type SystemExecutor interface {
	RunSimple(command string, args ...string) system.CommandResult
	DetectCPUInfo() (*system.CPUInfo, error)
}

// ChrootSession represents a persistent chroot session that can execute
// multiple commands while maintaining state (e.g., /tmp directory contents).
// Sessions must be explicitly closed when done to clean up resources.
type ChrootSession interface {
	// Exec executes a command in the chroot session
	Exec(command string) error

	// ExecWithOutput executes a command and returns its output
	ExecWithOutput(command string) (string, error)

	// ExecWithContext executes a command with context support for cancellation
	ExecWithContext(ctx context.Context, command string) error

	// ExecWithOutputAndContext executes a command with context and returns output
	ExecWithOutputAndContext(ctx context.Context, command string) (string, error)

	// Close terminates the session and cleans up resources
	Close() error
}

// ChrootExecutor abstracts chroot operations
type ChrootExecutor interface {
	ChrootExec(logPath, mountPoint, command string, args ...string) error
	ChrootExecWithOutput(logPath, mountPoint, command string, args ...string) (string, error)
	ChrootExecWithOutputAndContext(ctx context.Context, logPath, mountPoint, command string, args ...string) (string, error)
	ChrootSystemctl(logPath, mountPoint, action, service string) error
	ChrootExecWithStdin(logPath, mountPoint, command, stdin string) error
	ChrootExecWithContext(ctx context.Context, logPath, mountPoint, command string) error
	ChrootPacman(logPath, mountPoint, operation string, packages ...string) error
	DownloadAndInstallPackages(logPath, chrootPath string, urls ...string) error

	// BeginSession starts a persistent chroot session
	BeginSession(logPath, mountPoint string) (ChrootSession, error)

	// BeginSessionWithContext starts a persistent chroot session with context support
	BeginSessionWithContext(ctx context.Context, logPath, mountPoint string) (ChrootSession, error)
}

// HTTPClient abstracts HTTP operations
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// ConfigSaver abstracts config save operation
type ConfigSaver interface {
	Save() error
}

// DefaultFileSystem is the concrete implementation using OS functions
type DefaultFileSystem struct{}

func (d *DefaultFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (d *DefaultFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (d *DefaultFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (d *DefaultFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (d *DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (d *DefaultFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (d *DefaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (d *DefaultFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (d *DefaultFileSystem) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

// DefaultCommandExecutor is the concrete implementation using os/exec
type DefaultCommandExecutor struct{}

func (d *DefaultCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// DefaultHTTPClient is the concrete implementation using net/http
type DefaultHTTPClient struct{}

func (d *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

// DefaultSystemExecutor is the concrete implementation using system functions
type DefaultSystemExecutor struct{}

func (d *DefaultSystemExecutor) RunSimple(command string, args ...string) system.CommandResult {
	return system.RunSimple(command, args...)
}

func (d *DefaultSystemExecutor) DetectCPUInfo() (*system.CPUInfo, error) {
	return system.DetectCPUInfo()
}

// DefaultChrootExecutor is the concrete implementation using chroot functions
type DefaultChrootExecutor struct{}

func (d *DefaultChrootExecutor) ChrootExec(logPath, mountPoint, command string, args ...string) error {
	return system.ChrootExec(logPath, mountPoint, command, args...)
}

func (d *DefaultChrootExecutor) ChrootExecWithOutput(logPath, mountPoint, command string, args ...string) (string, error) {
	return system.ChrootExecWithOutput(logPath, mountPoint, command, args...)
}

func (d *DefaultChrootExecutor) ChrootExecWithOutputAndContext(ctx context.Context, logPath, mountPoint, command string, args ...string) (string, error) {
	return system.ChrootExecWithOutputAndContext(ctx, logPath, mountPoint, command, args...)
}

func (d *DefaultChrootExecutor) ChrootSystemctl(logPath, mountPoint, action, service string) error {
	return system.ChrootSystemctl(logPath, mountPoint, action, service)
}

func (d *DefaultChrootExecutor) ChrootExecWithStdin(logPath, mountPoint, command, stdin string) error {
	return system.ChrootExecWithStdin(logPath, mountPoint, command, stdin)
}

func (d *DefaultChrootExecutor) ChrootExecWithContext(ctx context.Context, logPath, mountPoint, command string) error {
	return system.ChrootExecWithContext(ctx, logPath, mountPoint, command)
}

func (d *DefaultChrootExecutor) ChrootPacman(logPath, mountPoint, operation string, packages ...string) error {
	return system.ChrootPacman(logPath, mountPoint, operation, packages...)
}

func (d *DefaultChrootExecutor) DownloadAndInstallPackages(logPath, chrootPath string, urls ...string) error {
	return system.DownloadAndInstallPackages(logPath, chrootPath, urls...)
}

func (d *DefaultChrootExecutor) BeginSession(logPath, mountPoint string) (ChrootSession, error) {
	return system.BeginSession(logPath, mountPoint)
}

func (d *DefaultChrootExecutor) BeginSessionWithContext(ctx context.Context, logPath, mountPoint string) (ChrootSession, error) {
	return system.BeginSessionWithContext(ctx, logPath, mountPoint)
}
