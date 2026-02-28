package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func errNotFound(name string) error {
	return fmt.Errorf("%s: not found", name)
}

// chaoticMocks sets up the mock expectations that are always required because
// Chaotic-AUR is unconditionally enabled.
func chaoticMocks(mockChrExec *mocks.MockChrootExecutor, mockFS *mocks.MockFileSystem) {
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--recv-key", gomock.Any(), "--keyserver", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--lsign-key", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-U", "--noconfirm", gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-Sy", "--noconfirm").Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-S", "--noconfirm", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockFS.EXPECT().ReadFile("/mnt/etc/pacman.conf").Return([]byte("[core]\nInclude = /etc/pacman.d/mirrorlist\n"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile("/mnt/etc/pacman.conf", gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
}

func TestReposHandler_Handle_Minimal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	chaoticMocks(mockChrExec, mockFS)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:     "/mnt",
		EnableMultilib: false,
		AURHelper:      packages.AURHelperParu,
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
	if result.AURHelper != "paru" {
		t.Errorf("expected paru helper, got %s", result.AURHelper)
	}
}

func TestReposHandler_Handle_MultilibEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	mockFS.EXPECT().ReadFile("/mnt/etc/pacman.conf").Return(
		[]byte("[core]\n#[multilib]\n#Include = /etc/pacman.d/mirrorlist\n"), nil,
	).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	chaoticMocks(mockChrExec, mockFS)
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-S", "--noconfirm", "--needed", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:     "/mnt",
		EnableMultilib: true,
		AURHelper:      packages.AURHelperYay,
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

	// pacman.conf read: first call for CachyOS, rest for Chaotic
	mockFS.EXPECT().ReadFile("/mnt/etc/pacman.conf").Return(
		[]byte("[core]\nInclude = /etc/pacman.d/mirrorlist\n"), nil,
	).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// WriteFile: cachyos mirrorlist written once; pacman.conf written multiple times
	// (once for chaotic-aur, once for cachyos) — verify the final write has [cachyos]
	mockFS.EXPECT().WriteFile("/mnt/etc/pacman.d/cachyos-mirrorlist", gomock.Any(), gomock.Any()).Return(nil)
	pacmanConfWriteCount := 0
	mockFS.EXPECT().WriteFile("/mnt/etc/pacman.conf", gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			pacmanConfWriteCount++
			return nil
		},
	).AnyTimes()
	t.Cleanup(func() {
		if pacmanConfWriteCount == 0 {
			t.Error("expected pacman.conf to be written at least once")
		}
	})

	// CachyOS key ops
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--init").Return([]byte{}, nil)
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--list-keys", "F3B607488DB35A47").Return(nil, errNotFound("key"))
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--recv-keys",
		"F3B607488DB35A47", "--keyserver", "keyserver.ubuntu.com").Return([]byte{}, nil)
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), "/mnt", "pacman-key", "--lsign-key",
		"F3B607488DB35A47").Return([]byte{}, nil)

	// Chaotic AUR key ops (always enabled)
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--recv-key", gomock.Any(), "--keyserver", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman-key", "--lsign-key", gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-U", "--noconfirm", gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-Sy", "--noconfirm").Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), "pacman", "-S", "--noconfirm", gomock.Any()).Return([]byte{}, nil).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint:    "/mnt",
		AURHelper:     packages.AURHelperParu,
		KernelVariant: packages.KernelCachyOS,
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

	chaoticMocks(mockChrExec, mockFS)
	mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, errNotFound("extra.packages")).AnyTimes()

	handler := NewReposHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.SetupRepositoriesCommand{
		MountPoint: "/mnt",
		AURHelper:  packages.AURHelperParu,
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
