package phases

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces/mocks"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
	"go.uber.org/mock/gomock"
)

// TestConfigPhasePreCheck tests configuration prerequisites
func TestConfigPhasePreCheck(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		username    string
		userPass    string
		timezone    string
		mountResult system.CommandResult
		wantErr     bool
		errContains string
	}{
		{
			name:        "All required fields present - /mnt mounted",
			hostname:    "archup-machine",
			username:    "archuser",
			userPass:    "password123",
			timezone:    "UTC",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     false,
		},
		{
			name:        "Missing hostname",
			hostname:    "",
			username:    "archuser",
			userPass:    "password123",
			timezone:    "UTC",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     true,
			errContains: "hostname is required",
		},
		{
			name:        "Missing username",
			hostname:    "archup-machine",
			username:    "",
			userPass:    "password123",
			timezone:    "UTC",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     true,
			errContains: "username is required",
		},
		{
			name:        "Missing user password",
			hostname:    "archup-machine",
			username:    "archuser",
			userPass:    "",
			timezone:    "UTC",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     true,
			errContains: "user password is required",
		},
		{
			name:        "Missing timezone",
			hostname:    "archup-machine",
			username:    "archuser",
			userPass:    "password123",
			timezone:    "",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     true,
			errContains: "timezone is required",
		},
		{
			name:        "/mnt not mounted",
			hostname:    "archup-machine",
			username:    "archuser",
			userPass:    "password123",
			timezone:    "UTC",
			mountResult: system.CommandResult{ExitCode: 32},
			wantErr:     true,
			errContains: "is not mounted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)
			mockChrExec := mocks.NewMockChrootExecutor(ctrl)

			mockSysExec.EXPECT().RunSimple("mountpoint", "-q", config.PathMnt).Return(tt.mountResult).Times(1)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			cfg.Hostname = tt.hostname
			cfg.Username = tt.username
			cfg.UserPassword = tt.userPass
			cfg.Timezone = tt.timezone

			phase := NewConfigPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

			err = phase.PreCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("PreCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("PreCheck() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestConfigPhasePostCheck tests configuration validation
func TestConfigPhasePostCheck(t *testing.T) {
	tests := []struct {
		name                    string
		setupMocks              func(*mocks.MockFileSystem)
		wantErr                 bool
		errContains             string
	}{
		{
			name: "All configuration files exist - postcheck passes",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Stat(config.PathMntEtcHostname).Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				mockFS.EXPECT().Stat(config.PathMntEtcLocaleConf).Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				zramPath := filepath.Join(config.PathMntEtcSystemd, config.FileZramGenerator)
				mockFS.EXPECT().Stat(zramPath).Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Hostname file missing - postcheck fails",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Stat(config.PathMntEtcHostname).Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(true).Times(1)
			},
			wantErr:     true,
			errContains: "hostname file was not created",
		},
		{
			name: "Locale.conf file missing - postcheck fails",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Stat(config.PathMntEtcHostname).Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				mockFS.EXPECT().Stat(config.PathMntEtcLocaleConf).Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(true).Times(1)
			},
			wantErr:     true,
			errContains: "locale.conf was not created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)
			mockChrExec := mocks.NewMockChrootExecutor(ctrl)

			tt.setupMocks(mockFS)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			cfg.ConfigPath = filepath.Join(tmpDir, "config")

			phase := NewConfigPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

			err = phase.PostCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstring(errMsg, tt.errContains) {
					t.Errorf("PostCheck() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestConfigPhaseExecute tests the Execute method
func TestConfigPhaseExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)

	// Setup successful calls for all configuration steps
	mockChrExec.EXPECT().ChrootExec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootExecWithStdin(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	cfg.Hostname = "archup"
	cfg.Username = "archuser"
	cfg.UserPassword = "password"
	cfg.Timezone = "UTC"
	cfg.Locale = "en_US"
	cfg.Keymap = "us"

	phase := NewConfigPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

	progressChan := make(chan ProgressUpdate, 100)
	result := phase.Execute(progressChan)

	if !result.Success {
		t.Errorf("Execute() Success = %v, want true", result.Success)
	}
	if result.Message != "Configuration complete" {
		t.Errorf("Execute() Message = %q, want 'Configuration complete'", result.Message)
	}
}

// TestConfigPhaseRollback tests the Rollback method
func TestConfigPhaseRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewConfigPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

	// ConfigPhase Rollback should be no-op
	err = phase.Rollback()
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}
}

// TestConfigPhaseCanSkip tests the CanSkip method
func TestConfigPhaseCanSkip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewConfigPhase(cfg, log, mockFS, mockSysExec, mockChrExec)

	if phase.CanSkip() {
		t.Error("CanSkip() = true, want false")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct{}

func (m mockFileInfo) Name() string       { return "mock" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }
