package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileRepository_NewFileRepository(t *testing.T) {
	tmpDir := t.TempDir()

	repo, err := NewFileRepository(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	if repo.basePath != tmpDir {
		t.Errorf("expected basePath %s, got %s", tmpDir, repo.basePath)
	}
}

func TestFileRepository_NewFileRepository_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "path")

	repo, err := NewFileRepository(nestedPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	// Verify directory was created
	info, err := os.Stat(nestedPath)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestFileRepository_Save_Success(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()
	state := "test installation state"

	err := repo.Save(ctx, "install-001", state)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "install-001.state")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected state file to be created")
	}
}

func TestFileRepository_Save_Content(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()
	state := "test installation state with content"

	repo.Save(ctx, "install-001", state)

	// Verify file content
	filePath := filepath.Join(tmpDir, "install-001.state")
	content, _ := os.ReadFile(filePath)

	if string(content) != state {
		t.Errorf("expected %s, got %s", state, string(content))
	}
}

func TestFileRepository_Save_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()

	// Save first state
	repo.Save(ctx, "install-001", "first state")

	// Overwrite with second state
	secondState := "second state"
	repo.Save(ctx, "install-001", secondState)

	// Verify file content is updated
	filePath := filepath.Join(tmpDir, "install-001.state")
	content, _ := os.ReadFile(filePath)

	if string(content) != secondState {
		t.Errorf("expected %s, got %s", secondState, string(content))
	}
}

func TestFileRepository_Load_Success(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	// Save first
	ctx := context.Background()
	state := "test installation state"
	repo.Save(ctx, "install-001", state)

	// Load
	loaded, err := repo.Load(ctx, "install-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loaded != state {
		t.Errorf("expected %s, got %s", state, loaded)
	}
}

func TestFileRepository_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()
	_, err := repo.Load(ctx, "nonexistent")

	if err == nil {
		t.Fatal("expected error for nonexistent installation")
	}
}

func TestFileRepository_Exists_True(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	// Save first
	ctx := context.Background()
	repo.Save(ctx, "install-001", "state")

	// Check exists
	exists, err := repo.Exists(ctx, "install-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected installation to exist")
	}
}

func TestFileRepository_Exists_False(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()
	exists, err := repo.Exists(ctx, "nonexistent")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected installation to not exist")
	}
}

func TestFileRepository_Save_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Save(ctx, "install-001", "state")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestFileRepository_Load_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	// Save first
	ctx := context.Background()
	repo.Save(ctx, "install-001", "state")

	// Load with cancelled context
	ctx2, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Load(ctx2, "install-001")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestFileRepository_Exists_ContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Exists(ctx, "install-001")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestFileRepository_MultipleInstallations(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx := context.Background()

	// Save multiple installations
	repo.Save(ctx, "install-001", "state 1")
	repo.Save(ctx, "install-002", "state 2")
	repo.Save(ctx, "install-003", "state 3")

	// Load each one
	state1, _ := repo.Load(ctx, "install-001")
	state2, _ := repo.Load(ctx, "install-002")
	state3, _ := repo.Load(ctx, "install-003")

	if state1 != "state 1" {
		t.Errorf("expected 'state 1', got %s", state1)
	}
	if state2 != "state 2" {
		t.Errorf("expected 'state 2', got %s", state2)
	}
	if state3 != "state 3" {
		t.Errorf("expected 'state 3', got %s", state3)
	}
}

func TestFileRepository_Save_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepository(tmpDir)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give it time to timeout
	time.Sleep(10 * time.Millisecond)

	err := repo.Save(ctx, "install-001", "state")
	if err == nil {
		t.Fatal("expected error due to context timeout")
	}
}
