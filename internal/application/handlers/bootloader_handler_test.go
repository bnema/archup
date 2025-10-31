package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestBootloaderHandler_Handle_Limine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewBootloaderHandler(mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.BootloaderType != "Limine" {
		t.Errorf("expected bootloader type Limine, got %s", result.BootloaderType)
	}

	if result.Timeout != 5 {
		t.Errorf("expected timeout 5, got %d", result.Timeout)
	}
}

func TestBootloaderHandler_Handle_SystemdBoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewBootloaderHandler(mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeSystemdBoot,
		TimeoutSeconds: 10,
		Branding:       "ArchUp",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.BootloaderType != "systemd-boot" {
		t.Errorf("expected bootloader type systemd-boot, got %s", result.BootloaderType)
	}
}

func TestBootloaderHandler_Handle_InvalidTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewBootloaderHandler(mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 700, // Too large
		Branding:       "ArchUp",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid timeout")
	}

	if result.Success {
		t.Error("expected failure for invalid timeout")
	}
}

func TestBootloaderHandler_Handle_InvalidBranding(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewBootloaderHandler(mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "", // Empty branding
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for empty branding")
	}

	if result.Success {
		t.Error("expected failure for empty branding")
	}
}
