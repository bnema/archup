package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileSystem_Stat(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected filename test.txt, got %s", info.Name())
	}
}

func TestLocalFileSystem_StatNotFound(t *testing.T) {
	fs := NewLocalFileSystem()

	_, err := fs.Stat("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLocalFileSystem_ReadFile(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("hello world")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	content, err := fs.ReadFile(testFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %s", string(content))
	}
}

func TestLocalFileSystem_ReadFileNotFound(t *testing.T) {
	fs := NewLocalFileSystem()

	_, err := fs.ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLocalFileSystem_WriteFile(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err := fs.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %s", string(data))
	}
}

func TestLocalFileSystem_WriteFileOverwrite(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Write first time
	if err := fs.WriteFile(testFile, []byte("first"), 0644); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Overwrite
	if err := fs.WriteFile(testFile, []byte("second"), 0644); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(testFile)
	if string(data) != "second" {
		t.Errorf("expected 'second', got %s", string(data))
	}
}

func TestLocalFileSystem_Create(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	file, err := fs.Create(testFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer file.Close()

	n, err := file.Write([]byte("test content"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if n != len("test content") {
		t.Errorf("expected to write 12 bytes, got %d", n)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("expected file to exist")
	}
}

func TestLocalFileSystem_CreateTruncates(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create file with content
	if err := os.WriteFile(testFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create again should truncate
	file, err := fs.Create(testFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	file.Close()

	// Verify file is empty
	info, _ := os.Stat(testFile)
	if info.Size() != 0 {
		t.Errorf("expected file size 0, got %d", info.Size())
	}
}

func TestLocalFileSystem_Chmod(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := fs.Chmod(testFile, 0755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info, _ := os.Stat(testFile)
	if info.Mode()&0755 != 0755 {
		t.Errorf("expected mode 0755, got %o", info.Mode()&0777)
	}
}

func TestLocalFileSystem_MkdirAll(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "a", "b", "c")
	err := fs.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestLocalFileSystem_RemoveAll(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	// Create directory with files
	testDir := filepath.Join(tmpDir, "test_dir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := fs.RemoveAll(testDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("expected directory to be removed")
	}
}

func TestLocalFileSystem_Exists_FileExists(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	exists, err := fs.Exists(testFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected file to exist")
	}
}

func TestLocalFileSystem_Exists_FileNotExists(t *testing.T) {
	fs := NewLocalFileSystem()

	exists, err := fs.Exists("/nonexistent/file.txt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected file to not exist")
	}
}

func TestLocalFileSystem_Exists_DirectoryExists(t *testing.T) {
	fs := NewLocalFileSystem()
	tmpDir := t.TempDir()

	exists, err := fs.Exists(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected directory to exist")
	}
}
