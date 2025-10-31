package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestConfigureSystemHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewConfigureSystemHandler(mockChrExec, mockLogger)

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "myarch",
		Timezone:     "UTC",
		Locale:       "en_US.UTF-8",
		Keymap:       "us",
		Username:     "testuser",
		UserShell:    "/bin/bash",
		UserPassword: "testpass123",
		RootPassword: "rootpass456",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	if result.Hostname != "myarch" {
		t.Errorf("expected hostname myarch, got %s", result.Hostname)
	}

	if result.Timezone != "UTC" {
		t.Errorf("expected timezone UTC, got %s", result.Timezone)
	}

	if result.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", result.Username)
	}
}

func TestConfigureSystemHandler_Handle_InvalidHostname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockChrExec, mockLogger)

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "", // Empty hostname
		Timezone:     "UTC",
		Locale:       "en_US.UTF-8",
		Keymap:       "us",
		Username:     "testuser",
		UserShell:    "/bin/bash",
		UserPassword: "testpass123",
		RootPassword: "rootpass456",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid hostname")
	}

	if result.Success {
		t.Error("expected failure for invalid hostname")
	}
}

func TestConfigureSystemHandler_Handle_InvalidUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockChrExec, mockLogger)

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "myarch",
		Timezone:     "UTC",
		Locale:       "en_US.UTF-8",
		Keymap:       "us",
		Username:     "", // Empty username
		UserShell:    "/bin/bash",
		UserPassword: "testpass123",
		RootPassword: "rootpass456",
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for invalid username")
	}

	if result.Success {
		t.Error("expected failure for invalid username")
	}
}

func TestConfigureSystemHandler_Handle_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockChrExec, mockLogger)

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "myarch",
		Timezone:     "UTC",
		Locale:       "en_US.UTF-8",
		Keymap:       "us",
		Username:     "testuser",
		UserShell:    "/bin/bash",
		UserPassword: "testpass123",
		RootPassword: "testpass123", // Same as user password
	}

	result, err := handler.Handle(context.Background(), cmd)

	if err == nil {
		t.Error("expected error for same passwords")
	}

	if result.Success {
		t.Error("expected failure for same passwords")
	}
}
