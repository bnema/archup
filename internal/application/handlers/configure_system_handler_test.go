package handlers

import (
	"context"
	"testing"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestXkbLayoutFromKeymap(t *testing.T) {
	tests := []struct {
		keymap   string
		expected string
	}{
		{"us", "us"},
		{"fr", "fr"},
		{"de", "de"},
		{"de-nodeadkeys", "de"},
		{"fr-latin9", "fr"},
		{"", ""},
	}
	for _, tt := range tests {
		got := xkbLayoutFromKeymap(tt.keymap)
		if got != tt.expected {
			t.Errorf("xkbLayoutFromKeymap(%q) = %q, want %q", tt.keymap, got, tt.expected)
		}
	}
}

func TestConfigureSystemHandler_Handle_FrenchKeymap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChrootWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	// Expect 00-keyboard.conf to be written with "fr" layout
	x11ConfPath := "/mnt/etc/X11/xorg.conf.d/00-keyboard.conf"
	mockFS.EXPECT().WriteFile(gomock.Eq(x11ConfPath), gomock.Any(), gomock.Any()).DoAndReturn(
		func(path string, content []byte, perm interface{}) error {
			if !containsString(string(content), `Option "XkbLayout" "fr"`) {
				t.Errorf("expected XkbLayout fr in 00-keyboard.conf, got: %s", string(content))
			}
			return nil
		},
	)
	// All other WriteFile calls
	mockFS.EXPECT().WriteFile(gomock.Not(gomock.Eq(x11ConfPath)), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)

	cmd := commands.ConfigureSystemCommand{
		MountPoint:   "/mnt",
		Hostname:     "myarch",
		Timezone:     "Europe/Paris",
		Locale:       "fr_FR.UTF-8",
		Keymap:       "fr",
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
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfigureSystemHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogPath().Return("/var/log/archup-install.log").AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChroot(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockChrExec.EXPECT().ExecuteInChrootWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	handler := NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)

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

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)

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

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)

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

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	handler := NewConfigureSystemHandler(mockFS, mockChrExec, mockLogger)

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
