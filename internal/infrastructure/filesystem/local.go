package filesystem

import (
	"fmt"
	"os"
)

// LocalFileSystem implements the FileSystem port using the OS filesystem
type LocalFileSystem struct{}

// NewLocalFileSystem creates a new local filesystem adapter
func NewLocalFileSystem() *LocalFileSystem {
	return &LocalFileSystem{}
}

// Stat returns file info for the given path
func (lfs *LocalFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile reads entire file content
func (lfs *LocalFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to a file with the given permissions
func (lfs *LocalFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Create creates or truncates a file
func (lfs *LocalFileSystem) Create(name string) (LocalFile, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &localFileHandle{file: file}, nil
}

// Chmod changes file permissions
func (lfs *LocalFileSystem) Chmod(name string, perm os.FileMode) error {
	return os.Chmod(name, perm)
}

// MkdirAll creates directory path with parents
func (lfs *LocalFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// RemoveAll removes a file or directory recursively
func (lfs *LocalFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Exists checks if a file/directory exists
func (lfs *LocalFileSystem) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check file existence: %w", err)
}

// LocalFile interface type for file operations
type LocalFile interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

// localFileHandle wraps os.File to implement the File port
type localFileHandle struct {
	file *os.File
}

// Read reads up to len(b) bytes from the file
func (lfh *localFileHandle) Read(b []byte) (n int, err error) {
	return lfh.file.Read(b)
}

// Write writes len(b) bytes to the file
func (lfh *localFileHandle) Write(b []byte) (n int, err error) {
	return lfh.file.Write(b)
}

// Close closes the file
func (lfh *localFileHandle) Close() error {
	return lfh.file.Close()
}
