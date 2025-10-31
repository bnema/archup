package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestPostInstallHandler_Handle_NoScripts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

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

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

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

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

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

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockScriptExec := mocks.NewMockScriptExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewPostInstallHandler(mockChrExec, mockScriptExec, mockLogger)

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
