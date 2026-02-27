package handlers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func errNotFound(name string) error {
	return fmt.Errorf("%s: not found", name)
}

func TestReposHandler_Handle_Minimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// extra.packages read fails gracefully (no file present)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  false,
		EnableChaotic:   false,
		AURHelper:       packages.AURHelperParu,
		AdditionalRepos: []string{},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Multilib {
		t.Error("expected multilib to be disabled")
	}
	if result.Chaotic {
		t.Error("expected chaotic to be disabled")
	}
	if result.AURHelper != "paru" {
		t.Errorf("expected paru helper, got %s", result.AURHelper)
	}
}

func TestReposHandler_Handle_AllEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	mockFS.EXPECT().ReadFile(gomock.Any()).Return(
		[]byte("[core]\n#[multilib]\n#Include = /etc/pacman.d/mirrorlist\n"), nil,
	).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Chaotic keyring key ops (chaotic uses --recv-key singular, pacman -U for keyring)
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--recv-key", gomock.Any(), "--keyserver", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--lsign-key", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-U", "--noconfirm", gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-Sy", "--noconfirm").Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-S", "--noconfirm", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-S", "--noconfirm", "--needed", gomock.Any()).Return([]byte{}, nil).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:      "/mnt",
		EnableMultilib:  true,
		EnableChaotic:   true,
		AURHelper:       packages.AURHelperYay,
		AdditionalRepos: []string{},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if !result.Multilib {
		t.Error("expected multilib to be enabled")
	}
	if !result.Chaotic {
		t.Error("expected chaotic to be enabled")
	}
	if result.AURHelper != "yay" {
		t.Errorf("expected yay helper, got %s", result.AURHelper)
	}
}

func TestReposHandler_Handle_CachyOSKernel_AddsCachyOSRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// extra.packages read fails gracefully; pacman.conf read succeeds
	mockFS.EXPECT().ReadFile("/mnt/etc/pacman.conf").Return(
		[]byte("[core]\nInclude = /etc/pacman.d/mirrorlist\n"), nil,
	)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	// MkdirAll for /mnt/etc/pacman.d
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)

	// WriteFile: mirrorlist + pacman.conf
	mockFS.EXPECT().WriteFile("/mnt/etc/pacman.d/cachyos-mirrorlist", gomock.Any(), gomock.Any()).Return(nil)
	mockFS.EXPECT().WriteFile("/mnt/etc/pacman.conf", gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			if !strings.Contains(string(data), "[cachyos]") {
				t.Errorf("expected pacman.conf to contain [cachyos], got:\n%s", string(data))
			}
			return nil
		},
	)

	// pacman-key --init
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--init").Return([]byte{}, nil)
	// --list-keys: key not present → triggers --recv-keys
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--list-keys", "F3B607488DB35A47").Return(nil, errNotFound("key"))
	// --recv-keys
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--recv-keys",
		"F3B607488DB35A47", "--keyserver", "keyserver.ubuntu.com").Return([]byte{}, nil)
	// --lsign-key
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--lsign-key",
		"F3B607488DB35A47").Return([]byte{}, nil)

	// No pacman -Sy here: CachyOS-only (no multilib, no chaotic) — sync is skipped in repos phase
	// (chroot sync only runs when multilib or chaotic are enabled)

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:     "/mnt",
		EnableMultilib: false,
		EnableChaotic:  false,
		AURHelper:      packages.AURHelperParu,
		KernelVariant:  packages.KernelCachyOS,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestReposHandler_Handle_WithAdditionalRepos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:     "/mnt",
		EnableMultilib: false,
		EnableChaotic:  false,
		AURHelper:      packages.AURHelperParu,
		AdditionalRepos: []string{
			"https://example.com/archup",
			"https://custom.repo/packages",
		},
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}
