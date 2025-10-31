package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestPartitionHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPartitionHandler(mockExec, mockLogger)

	cmd := commands.PartitionDiskCommand{
		TargetDisk:         "/dev/sda",
		RootSizeGB:         50,
		BootSizeGB:         1,
		EncryptionType:     disk.EncryptionTypeNone,
		EncryptionPassword: "",
		FilesystemType:     disk.FilesystemExt4,
		WipeDisks:          true,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.ErrorDetail)
	}

	if result.TargetDisk != "/dev/sda" {
		t.Errorf("expected target disk /dev/sda, got %s", result.TargetDisk)
	}

	if len(result.Partitions) != 2 {
		t.Errorf("expected 2 partitions, got %d", len(result.Partitions))
	}

	// Check boot partition
	if len(result.Partitions) > 0 && result.Partitions[0].MountPoint != "/boot/efi" {
		t.Errorf("expected first partition mount point /boot/efi, got %s", result.Partitions[0].MountPoint)
	}

	// Check root partition
	if len(result.Partitions) > 1 && result.Partitions[1].MountPoint != "/" {
		t.Errorf("expected second partition mount point /, got %s", result.Partitions[1].MountPoint)
	}
}

func TestPartitionHandler_Handle_InvalidDisk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewPartitionHandler(mockExec, mockLogger)

	cmd := commands.PartitionDiskCommand{
		TargetDisk:         "",
		RootSizeGB:         50,
		BootSizeGB:         1,
		EncryptionType:     disk.EncryptionTypeNone,
		FilesystemType:     disk.FilesystemExt4,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid disk")
	}

	if result.Success {
		t.Error("expected result to indicate failure")
	}
}

func TestPartitionHandler_Handle_RootPartitionTooSmall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewPartitionHandler(mockExec, mockLogger)

	cmd := commands.PartitionDiskCommand{
		TargetDisk:         "/dev/sda",
		RootSizeGB:         1, // Too small
		BootSizeGB:         1,
		EncryptionType:     disk.EncryptionTypeNone,
		FilesystemType:     disk.FilesystemExt4,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for root partition too small")
	}

	if result.Success {
		t.Error("expected result to indicate failure")
	}
}
