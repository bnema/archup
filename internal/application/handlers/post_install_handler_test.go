package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func newMockResponse(ctrl *gomock.Controller, statusCode int, body []byte) ports.Response {
	resp := mocks.NewMockResponse(ctrl)
	resp.EXPECT().StatusCode().Return(statusCode).AnyTimes()
	resp.EXPECT().Body().Return(body).AnyTimes()
	resp.EXPECT().Close().Return(nil).AnyTimes()
	return resp
}

func TestPostInstallHandler_Handle_NoScripts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.TasksRun) != 0 {
		t.Errorf("expected no tasks run, got %d", len(result.TasksRun))
	}
}

func TestPostInstallHandler_Handle_WithScripts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("content"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Chmod(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: true,
		PlymouthTheme:      "",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.TasksRun) != 1 {
		t.Errorf("expected 1 task run, got %d", len(result.TasksRun))
	}

	if result.TasksRun[0] != "post-boot-scripts" {
		t.Errorf("expected task 'post-boot-scripts', got %s", result.TasksRun[0])
	}
}

func TestPostInstallHandler_Handle_WithPlymouth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "spinner",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.TasksRun) != 1 {
		t.Errorf("expected 1 task run, got %d", len(result.TasksRun))
	}

	if result.TasksRun[0] != "plymouth-theme-spinner" {
		t.Errorf("expected task 'plymouth-theme-spinner', got %s", result.TasksRun[0])
	}
}

func TestPostInstallHandler_Handle_WithDankLinux(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Eq("/mnt/var/lib/archup-install-danklinux"), gomock.Any(), gomock.Any()).Return(nil)
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
		InstallDankLinux:   true,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.TasksRun) != 1 {
		t.Errorf("expected 1 task run, got %d", len(result.TasksRun))
	}

	if result.TasksRun[0] != "dank-linux-flag" {
		t.Errorf("expected task 'dank-linux-flag', got %s", result.TasksRun[0])
	}
}

func TestPostInstallHandler_Handle_TunesPacman(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	pacmanContent := "#Color\n#ParallelDownloads = 5\n#ILoveCandy\n"

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	// ReadFile for pacman.conf returns commented options; other reads return generic content
	mockFS.EXPECT().ReadFile(gomock.Eq("/mnt/etc/pacman.conf")).Return([]byte(pacmanContent), nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()

	// Capture the WriteFile call for pacman.conf and assert content
	mockFS.EXPECT().WriteFile(gomock.Eq("/mnt/etc/pacman.conf"), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			content := string(data)
			if !strings.Contains(content, "Color") {
				t.Errorf("expected tuned pacman.conf to contain 'Color', got: %s", content)
			}
			if strings.Contains(content, "#Color") {
				t.Errorf("expected '#Color' to be uncommented in pacman.conf")
			}
			if !strings.Contains(content, "ParallelDownloads = 5") {
				t.Errorf("expected tuned pacman.conf to contain 'ParallelDownloads = 5'")
			}
			if !strings.Contains(content, "ILoveCandy") {
				t.Errorf("expected tuned pacman.conf to contain 'ILoveCandy'")
			}
			return nil
		},
	).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestPostInstallHandler_Handle_InstallsLimineHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	hooksDir := "/mnt/etc/pacman.d/hooks"
	hookFile := hooksDir + "/limine-update.hook"
	targetDisk := "/dev/sda"
	hookTemplate := "Exec = /usr/bin/limine bios-install DISK_PLACEHOLDER"

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte(hookTemplate)), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	hookWritten := false
	mockFS.EXPECT().WriteFile(gomock.Eq(hookFile), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, data []byte, perm interface{}) error {
			hookWritten = true
			content := string(data)
			if strings.Contains(content, "DISK_PLACEHOLDER") {
				t.Errorf("hook content should not contain DISK_PLACEHOLDER, got: %s", content)
			}
			if !strings.Contains(content, targetDisk) {
				t.Errorf("hook content should contain target disk %s, got: %s", targetDisk, content)
			}
			return nil
		},
	).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
		TargetDisk:         targetDisk,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if !hookWritten {
		t.Error("expected limine hook file to be written")
	}
}

func TestPostInstallHandler_Handle_EnablesSnapperSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	pacmanInstalled := false
	snapperTimerEnabled := false

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()

	// Track pacman install of limine-snapper-sync (specific expectation first)
	mockChrExec.EXPECT().ExecuteInChroot(
		gomock.Any(), gomock.Eq("/mnt"), gomock.Eq("pacman"),
		gomock.Eq("-S"), gomock.Eq("--noconfirm"), gomock.Eq("--needed"), gomock.Eq("limine-snapper-sync"),
	).DoAndReturn(func(ctx context.Context, mountPoint, command string, args ...string) ([]byte, error) {
		pacmanInstalled = true
		return []byte{}, nil
	}).AnyTimes()
	// Catch-all for any other ExecuteInChroot calls
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()

	// Track systemctl enable for limine-snapper-sync.timer (specific expectation first)
	mockChrExec.EXPECT().ChrootSystemctl(
		gomock.Any(), gomock.Any(), gomock.Eq("/mnt"), gomock.Eq("enable"), gomock.Eq("limine-snapper-sync.timer"),
	).DoAndReturn(func(ctx context.Context, logPath, chrootPath string, args ...string) error {
		snapperTimerEnabled = true
		return nil
	}).AnyTimes()
	// Catch-all for any other ChrootSystemctl calls
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if !pacmanInstalled {
		t.Error("expected pacman to install limine-snapper-sync")
	}

	if !snapperTimerEnabled {
		t.Error("expected limine-snapper-sync.timer to be enabled via systemctl")
	}
}

func TestPostInstallHandler_Handle_Everything(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("content"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Chmod(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: true,
		PlymouthTheme:      "bgrt",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.TasksRun) != 2 {
		t.Errorf("expected 2 tasks run, got %d", len(result.TasksRun))
	}

	// Check that both tasks are in the results
	tasksFound := make(map[string]bool)
	for _, task := range result.TasksRun {
		tasksFound[task] = true
	}

	if !tasksFound["post-boot-scripts"] {
		t.Error("expected post-boot-scripts task")
	}

	if !tasksFound["plymouth-theme-bgrt"] {
		t.Error("expected plymouth-theme-bgrt task")
	}
}

func TestPostInstallHandler_Handle_RunsVerification(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Stat returns nil (files exist) for the checked paths
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
		Encrypted:          false,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if len(result.VerificationWarnings) != 0 {
		t.Errorf("expected no verification warnings, got: %v", result.VerificationWarnings)
	}
}

func TestPostInstallHandler_Handle_VerificationWarnsOnMissingFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("graphics: yes"), nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(newMockResponse(ctrl, http.StatusOK, []byte("content")), nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// limine.conf is missing; all other Stat calls succeed
	mockFS.EXPECT().Stat(gomock.Eq("/mnt/boot/limine.conf")).Return(nil, fmt.Errorf("not found"))
	mockFS.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger, "https://raw.githubusercontent.com/bnema/archup/dev")

	cmd := commands.PostInstallCommand{
		MountPoint:         "/mnt",
		Username:           "testuser",
		RunPostBootScripts: false,
		PlymouthTheme:      "",
		Encrypted:          false,
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success even with verification warnings")
	}

	if len(result.VerificationWarnings) == 0 {
		t.Error("expected verification warnings for missing limine.conf")
	}

	found := false
	for _, w := range result.VerificationWarnings {
		if strings.Contains(w, "limine.conf") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about limine.conf, got: %v", result.VerificationWarnings)
	}
}
