package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestPreflightHandler_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	// Expect basic logging calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Mock filesystem and executor responses
	mockFS.EXPECT().Exists(gomock.Any()).DoAndReturn(func(path string) (bool, error) {
		switch path {
		case "/etc/arch-release":
			return true, nil
		case "/sys/firmware/efi/efivars":
			return true, nil
		default:
			return false, nil
		}
	}).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "id", "-u").Return([]byte("0\n"), nil).Times(1)
	mockExec.EXPECT().Execute(gomock.Any(), "uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
	mockExec.EXPECT().Execute(gomock.Any(), "bootctl", "status").Return([]byte("Secure Boot: disabled"), nil).Times(2)
	mockExec.EXPECT().Execute(gomock.Any(), "ping", "-c", "1", "archlinux.org").Return([]byte{}, nil).Times(1)
	mockExec.EXPECT().Execute(gomock.Any(), "grep", "model name", "/proc/cpuinfo").Return([]byte("model name  : Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz\n"), nil).Times(1)

	handler := NewPreflightHandler(mockFS, mockExec, mockLogger)
	result, err := handler.Handle(context.Background(), commands.PreflightCommand{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.ChecksPassed {
		t.Error("expected checks to pass")
	}

	if result.SystemInfo.Architecture != "x86_64" {
		t.Errorf("expected architecture x86_64, got %s", result.SystemInfo.Architecture)
	}

	if !result.SystemInfo.IsUEFI {
		t.Error("expected UEFI boot to be detected")
	}
}

func TestPreflightHandler_NotRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).DoAndReturn(func(path string) (bool, error) {
		switch path {
		case "/etc/arch-release":
			return true, nil
		default:
			return false, nil
		}
	}).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "id", "-u").Return([]byte("1000\n"), nil).Times(1)
	mockExec.EXPECT().Execute(gomock.Any(), "bootctl", "status").Return([]byte("Secure Boot: disabled"), nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPreflightHandler(mockFS, mockExec, mockLogger)
	result, err := handler.Handle(context.Background(), commands.PreflightCommand{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ChecksPassed {
		t.Error("expected checks to fail when not running as root")
	}

	if len(result.CriticalErrors) == 0 {
		t.Error("expected critical errors to be recorded")
	}
}
