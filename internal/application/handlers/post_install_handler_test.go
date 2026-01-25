package handlers

import (
	"context"
	"net/http"
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
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger)

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

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger)

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

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger)

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

	handler := NewPostInstallHandler(mockFS, mockHTTP, mockChrExec, mockScriptExec, mockLogger)

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
