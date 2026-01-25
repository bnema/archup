package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestInstallBaseHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "pacstrap", "/mnt", "base", "linux-firmware", "linux-zen", "intel-ucode", "amd-ucode", "vim", "git").Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "genfstab", "-U", "/mnt").Return([]byte("# fstab"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBaseCommand{
		TargetDisk:       "/dev/sda",
		MountPoint:       "/mnt",
		Packages:         []string{"vim", "git"},
		KernelVariant:    packages.KernelZen,
		IncludeMicrocode: true,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.PackagesInstalled) < 4 {
		t.Errorf("expected at least 4 packages, got %d", len(result.PackagesInstalled))
	}

	// Check that linux-zen is in the list
	found := false
	for _, pkg := range result.PackagesInstalled {
		if pkg == "linux-zen" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected linux-zen kernel to be in packages list")
	}
}

func TestInstallBaseHandler_Handle_InvalidKernel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)

	cmd := commands.InstallBaseCommand{
		TargetDisk:    "/dev/sda",
		MountPoint:    "/mnt",
		KernelVariant: packages.KernelVariant(999), // Invalid kernel
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid kernel")
	}

	if result.Success {
		t.Error("expected failure for invalid kernel")
	}
}

func TestInstallBaseHandler_Handle_WithCustomPackages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockExec := mocks.NewMockCommandExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "pacstrap", "/mnt", "base", "linux-firmware", "linux", "neovim", "zsh", "tmux").Return([]byte{}, nil).AnyTimes()
	mockExec.EXPECT().Execute(gomock.Any(), "genfstab", "-U", "/mnt").Return([]byte("# fstab"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewInstallBaseHandler(mockFS, mockExec, mockChrExec, mockLogger)

	customPkgs := []string{"neovim", "zsh", "tmux"}
	cmd := commands.InstallBaseCommand{
		TargetDisk:       "/dev/sda",
		MountPoint:       "/mnt",
		Packages:         customPkgs,
		KernelVariant:    packages.KernelStable,
		IncludeMicrocode: false,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	// Check that custom packages are included
	for _, customPkg := range customPkgs {
		found := false
		for _, installedPkg := range result.PackagesInstalled {
			if installedPkg == customPkg {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected custom package %s to be installed", customPkg)
		}
	}
}
