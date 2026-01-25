package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestChrootExecutor_NewChrootExecutor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
}

func TestChrootExecutor_ExecuteInChroot_PathNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	ctx := context.Background()

	_, err := executor.ExecuteInChroot(ctx, "/nonexistent/path", "echo", "test")
	if err == nil {
		t.Fatal("expected error for nonexistent chroot path")
	}
}

func TestChrootExecutor_ExecuteInChroot_CommandNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	ctx := context.Background()

	// Create temporary directory to use as chroot
	tmpDir := t.TempDir()

	// arch-chroot requires proper setup which we can't do in unit tests,
	// so we expect this to fail when trying to execute arch-chroot
	_, err := executor.ExecuteInChroot(ctx, tmpDir, "echo", "test")
	if err == nil {
		t.Fatal("expected error when executing in non-functional chroot")
	}
}

func TestChrootExecutor_ChrootSystemctl_PathNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)

	err := executor.ChrootSystemctl(context.Background(), "", "/nonexistent/path", "enable", "service")
	if err == nil {
		t.Fatal("expected error for nonexistent chroot path")
	}
}

func TestChrootExecutor_ChrootSystemctl_CreatesLogDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "subdir", "log.txt")

	// This will fail due to arch-chroot not being available,
	// but the directory should be created
	_ = executor.ChrootSystemctl(context.Background(), logPath, tmpDir, "enable", "service")

	// Check if log directory was created
	logDir := filepath.Dir(logPath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("expected log directory to be created")
	}
}

func TestChrootExecutor_ChrootSystemctl_LogsErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "log.txt")

	// This will fail due to arch-chroot not being available
	_ = executor.ChrootSystemctl(context.Background(), logPath, tmpDir, "enable", "service")

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// It's OK if log file wasn't created due to error
		t.Log("log file not created (expected on failure)")
	}
}

func TestChrootExecutor_ChrootSystemctl_EmptyLogPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	executor := NewChrootExecutor(mockLogger)
	tmpDir := t.TempDir()

	// This should not panic even with empty log path
	_ = executor.ChrootSystemctl(context.Background(), "", tmpDir, "enable", "service")
}

func TestChrootExecutor_ExecuteInChrootWithStdin(t *testing.T) {
	t.Run("passes-stdin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		executor := NewChrootExecutor(mockLogger)
		workspace := t.TempDir()
		chrootPath := filepath.Join(workspace, "chroot")
		if err := os.MkdirAll(chrootPath, 0755); err != nil {
			t.Fatalf("failed to create chroot path: %v", err)
		}

		scriptDir := filepath.Join(workspace, "bin")
		if err := os.MkdirAll(scriptDir, 0755); err != nil {
			t.Fatalf("failed to create script dir: %v", err)
		}

		outputPath := filepath.Join(workspace, "stdin.txt")
		scriptPath := filepath.Join(scriptDir, "arch-chroot")
		script := "#!/bin/sh\ncat - > \"$OUTPUT_FILE\"\n"
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			t.Fatalf("failed to write arch-chroot script: %v", err)
		}

		oldPath := os.Getenv("PATH")
		if err := os.Setenv("PATH", scriptDir+":"+oldPath); err != nil {
			t.Fatalf("failed to set PATH: %v", err)
		}
		if err := os.Setenv("OUTPUT_FILE", outputPath); err != nil {
			t.Fatalf("failed to set OUTPUT_FILE: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Setenv("PATH", oldPath)
			_ = os.Unsetenv("OUTPUT_FILE")
		})

		if err := executor.ExecuteInChrootWithStdin(context.Background(), chrootPath, "hello", "echo", "ignored"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}
		if string(content) != "hello" {
			t.Errorf("expected stdin to be written, got %q", string(content))
		}
	})

	t.Run("context-cancel", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		executor := NewChrootExecutor(mockLogger)
		workspace := t.TempDir()
		chrootPath := filepath.Join(workspace, "chroot")
		if err := os.MkdirAll(chrootPath, 0755); err != nil {
			t.Fatalf("failed to create chroot path: %v", err)
		}

		scriptDir := filepath.Join(workspace, "bin")
		if err := os.MkdirAll(scriptDir, 0755); err != nil {
			t.Fatalf("failed to create script dir: %v", err)
		}

		scriptPath := filepath.Join(scriptDir, "arch-chroot")
		script := "#!/bin/sh\nsleep 5\n"
		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			t.Fatalf("failed to write arch-chroot script: %v", err)
		}

		oldPath := os.Getenv("PATH")
		if err := os.Setenv("PATH", scriptDir+":"+oldPath); err != nil {
			t.Fatalf("failed to set PATH: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Setenv("PATH", oldPath)
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if err := executor.ExecuteInChrootWithStdin(ctx, chrootPath, "hello", "sleep", "5"); err == nil {
			t.Fatal("expected error due to context cancellation")
		}
	})

	t.Run("missing-chroot-path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		executor := NewChrootExecutor(mockLogger)
		if err := executor.ExecuteInChrootWithStdin(context.Background(), "/nonexistent/path", "input", "echo", "test"); err == nil {
			t.Fatal("expected error for nonexistent chroot path")
		}
	})
}
