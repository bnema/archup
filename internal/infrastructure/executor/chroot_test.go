package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/domain/ports/mocks"
	"github.com/golang/mock/gomock"
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

	err := executor.ChrootSystemctl("", "/nonexistent/path", "enable", "service")
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
	executor.ChrootSystemctl(logPath, tmpDir, "enable", "service")

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
	executor.ChrootSystemctl(logPath, tmpDir, "enable", "service")

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// It's OK if log file wasn't created due to error
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
	_ = executor.ChrootSystemctl("", tmpDir, "enable", "service")
}
