package handlers

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

// limineTemplate is the minimal template content used in tests to avoid reading real files.
// It mirrors the structure of limine.conf.template for the purpose of testing placeholder replacement.
const limineTemplate = `timeout: {{TIMEOUT}}
default_entry: 2
quiet: yes
interface_branding: {{BRANDING}}
interface_branding_colour: cyan
graphics: yes
backdrop: 000000

/+{{KERNEL}}
comment: machine-id={{MACHINE_ID}}

    //{{KERNEL}}
    protocol: linux
    path: boot():/vmlinuz-{{KERNEL}}
    cmdline: {{KERNEL_PARAMS}}
    module_path: boot():/initramfs-{{KERNEL}}.img
{{FALLBACK_ENTRY}}
    //Snapshots
`

// setupCommonMocks wires the expectations shared across bootloader handler tests.
// Callers MUST register their own ReadFile and Stat expectations before calling configureLimine.
func setupCommonMocks(mockFS *mocks.MockFileSystem, mockExec *mocks.MockCommandExecutor, mockChrExec *mocks.MockChrootExecutor, mockLogger *mocks.MockLogger) {
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "blkid", "-s", "UUID", "-o", "value", gomock.Any()).Return([]byte("uuid"), nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "cp", gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
}

func TestBootloaderHandler_Handle_Limine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	setupCommonMocks(mockFS, mockExec, mockChrExec, mockLogger)
	// ReadFile: mkinitcpio.conf and machine-id return defaults; template returns limineTemplate
	mockFS.EXPECT().ReadFile(gomock.Any()).DoAndReturn(func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "limine.conf.template") {
			return []byte(limineTemplate), nil
		}
		return []byte("HOOKS=(base)\n"), nil
	}).AnyTimes()
	// Stat: fallback image absent (os.ErrNotExist so the warn branch is not taken)
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, os.ErrNotExist).AnyTimes()

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

// TestConfigureLimine_FallbackAbsent verifies that when the fallback initramfs image does not
// exist, the written limine.conf contains no "fallback" reference.
func TestConfigureLimine_FallbackAbsent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// blkid returns a UUID
	mockExec.EXPECT().Execute(gomock.Any(), "blkid", "-s", "UUID", "-o", "value", gomock.Any()).
		Return([]byte("test-uuid"), nil)

	// Template lookup: Exists returns true for the template path
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()

	// ReadFile returns the limine template for the template file, and machine-id otherwise
	mockFS.EXPECT().ReadFile(gomock.Any()).DoAndReturn(func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "limine.conf.template") {
			return []byte(limineTemplate), nil
		}
		return []byte("abc123\n"), nil
	}).AnyTimes()

	// Stat: fallback image does NOT exist (os.ErrNotExist so the warn branch is not taken)
	mockFS.EXPECT().Stat("/mnt/boot/initramfs-linux-fallback.img").Return(nil, os.ErrNotExist)

	// Capture the written config
	var writtenConfig string
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			if strings.HasSuffix(path, "limine.conf") {
				writtenConfig = string(data)
			}
			return nil
		},
	)

	handler := NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
	}

	err := handler.configureLimine(context.Background(), cmd, "linux")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if strings.Contains(writtenConfig, "fallback") {
		t.Errorf("expected no 'fallback' in limine.conf when fallback image is absent, got:\n%s", writtenConfig)
	}
}

// TestConfigureLimine_FallbackPresent verifies that when the fallback initramfs image exists,
// the written limine.conf contains the full fallback stanza.
func TestConfigureLimine_FallbackPresent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// blkid returns a UUID
	mockExec.EXPECT().Execute(gomock.Any(), "blkid", "-s", "UUID", "-o", "value", gomock.Any()).
		Return([]byte("test-uuid"), nil)

	// Template lookup
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()

	// ReadFile returns the limine template for the template file, and machine-id otherwise
	mockFS.EXPECT().ReadFile(gomock.Any()).DoAndReturn(func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "limine.conf.template") {
			return []byte(limineTemplate), nil
		}
		return []byte("abc123\n"), nil
	}).AnyTimes()

	// Stat: fallback image EXISTS (nil error)
	mockFS.EXPECT().Stat("/mnt/boot/initramfs-linux-fallback.img").Return(nil, nil)

	// Capture the written config
	var writtenConfig string
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			if strings.HasSuffix(path, "limine.conf") {
				writtenConfig = string(data)
			}
			return nil
		},
	)

	handler := NewBootloaderHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBootloaderCommand{
		MountPoint:     "/mnt",
		BootloaderType: bootloader.BootloaderTypeLimine,
		TimeoutSeconds: 5,
		Branding:       "ArchUp",
		KernelVariant:  packages.KernelStable,
		RootPartition:  "/dev/sda2",
		EncryptionType: disk.EncryptionTypeNone,
	}

	err := handler.configureLimine(context.Background(), cmd, "linux")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(writtenConfig, "initramfs-linux-fallback.img") {
		t.Errorf("expected 'initramfs-linux-fallback.img' in limine.conf when fallback image is present, got:\n%s", writtenConfig)
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
