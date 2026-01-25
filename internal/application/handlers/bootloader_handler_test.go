package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestBootloaderHandler_Handle_Limine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("HOOKS=(base)\n"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "blkid", "-s", "UUID", "-o", "value", gomock.Any()).Return([]byte("uuid"), nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "cp", gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "mkinitcpio", "-P").Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "efibootmgr", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()

	handler := NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
		EFIPartition:   "/dev/sda1",
		TargetDisk:     "/dev/sda",
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

func TestBootloaderHandler_Handle_InvalidTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 700, // Too large
		Branding:       "ArchUp",
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
		EFIPartition:   "/dev/sda1",
		TargetDisk:     "/dev/sda",
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

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "", // Empty branding
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
		EFIPartition:   "/dev/sda1",
		TargetDisk:     "/dev/sda",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for empty branding")
	}

	if result.Success {
		t.Error("expected failure for empty branding")
	}
}
