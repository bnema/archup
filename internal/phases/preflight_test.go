package phases

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces/mocks"
	"github.com/bnema/archup/internal/logger"
	"go.uber.org/mock/gomock"
)

// TestPreflightPhasePreCheck tests the PreCheck validation
func TestPreflightPhasePreCheck(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockFileSystem, *mocks.MockCommandExecutor)
		wantErr       bool
		errContains   string
	}{
		{
			name: "Arch Linux detection - file not found",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Not(nil)).Return(true).Times(1)
			},
			wantErr:     true,
			errContains: "must be running on Arch Linux",
		},
		{
			name: "Arch Linux detected - CachyOS derivative detected",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				// For derivatives, check all 4 since map iteration order isn't guaranteed
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).AnyTimes()
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).AnyTimes()
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).AnyTimes()
			},
			wantErr:     true,
			errContains: "CachyOS",
		},
		{
			name: "Arch Linux detected - vanilla Arch no derivatives",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(2)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
				mockFS.EXPECT().Stat("/sys/firmware/efi/efivars").Return(mockFileInfo{}, nil).Times(1)
				mockCmd.EXPECT().Execute("bootctl", "status").Return([]byte("Secure Boot: disabled\n"), nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "Vanilla Arch - CPU detection fails",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return(nil, fmt.Errorf("uname failed")).Times(1)
			},
			wantErr:     true,
			errContains: "failed to detect architecture",
		},
		{
			name: "Vanilla Arch - CPU not x86_64",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("armv7l\n"), nil).Times(1)
			},
			wantErr:     true,
			errContains: "must be x86_64",
		},
		{
			name: "x86_64 detected - UEFI not found",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
				mockFS.EXPECT().Stat("/sys/firmware/efi/efivars").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(true).Times(1)
			},
			wantErr:     true,
			errContains: "must be UEFI boot mode",
		},
		{
			name: "UEFI detected - Secure Boot enabled",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(2)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
				mockFS.EXPECT().Stat("/sys/firmware/efi/efivars").Return(mockFileInfo{}, nil).Times(1)
				mockCmd.EXPECT().Execute("bootctl", "status").Return([]byte("Secure Boot: enabled\n"), nil).Times(1)
			},
			wantErr:     true,
			errContains: "Secure Boot must be disabled",
		},
		{
			name: "Secure Boot disabled - all checks pass",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(2)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
				mockFS.EXPECT().Stat("/sys/firmware/efi/efivars").Return(mockFileInfo{}, nil).Times(1)
				mockCmd.EXPECT().Execute("bootctl", "status").Return([]byte("Secure Boot: disabled\n"), nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "bootctl command fails - precheck should still pass",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockCmd *mocks.MockCommandExecutor) {
				mockFS.EXPECT().Stat("/etc/arch-release").Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(2)
				// All derivatives should return not found
				mockFS.EXPECT().Stat("/etc/cachyos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/eos-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/garuda-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().Stat("/etc/manjaro-release").Return(nil, fmt.Errorf("not found")).Times(1)
				mockCmd.EXPECT().Execute("uname", "-m").Return([]byte("x86_64\n"), nil).Times(1)
				mockFS.EXPECT().Stat("/sys/firmware/efi/efivars").Return(mockFileInfo{}, nil).Times(1)
				mockCmd.EXPECT().Execute("bootctl", "status").Return(nil, fmt.Errorf("bootctl not available")).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockCmd := mocks.NewMockCommandExecutor(ctrl)

			tt.setupMocks(mockFS, mockCmd)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)

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

// TestPreflightPhaseSetDefaults tests setting default configuration values
func TestPreflightPhaseSetDefaults(t *testing.T) {
	tests := []struct {
		name               string
		initialHostname    string
		initialTimezone    string
		expectedHostname   string
		expectedTimezone   string
		expectedBootloader string
	}{
		{
			name:               "Empty hostname and timezone - use defaults",
			initialHostname:    "",
			initialTimezone:    "",
			expectedHostname:   "archup",
			expectedTimezone:   "UTC",
			expectedBootloader: config.BootloaderLimine,
		},
		{
			name:               "Custom hostname provided - keep custom",
			initialHostname:    "my-arch",
			initialTimezone:    "",
			expectedHostname:   "my-arch",
			expectedTimezone:   "UTC",
			expectedBootloader: config.BootloaderLimine,
		},
		{
			name:               "Custom timezone provided - keep custom",
			initialHostname:    "",
			initialTimezone:    "Europe/London",
			expectedHostname:   "archup",
			expectedTimezone:   "Europe/London",
			expectedBootloader: config.BootloaderLimine,
		},
		{
			name:               "Both custom - keep both",
			initialHostname:    "arch-desktop",
			initialTimezone:    "America/New_York",
			expectedHostname:   "arch-desktop",
			expectedTimezone:   "America/New_York",
			expectedBootloader: config.BootloaderLimine,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockCmd := mocks.NewMockCommandExecutor(ctrl)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			cfg.Hostname = tt.initialHostname
			cfg.Timezone = tt.initialTimezone

			phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)
			phase.SetDefaults()

			if cfg.Hostname != tt.expectedHostname {
				t.Errorf("Hostname = %q, want %q", cfg.Hostname, tt.expectedHostname)
			}
			if cfg.Timezone != tt.expectedTimezone {
				t.Errorf("Timezone = %q, want %q", cfg.Timezone, tt.expectedTimezone)
			}
			if cfg.Bootloader != tt.expectedBootloader {
				t.Errorf("Bootloader = %q, want %q", cfg.Bootloader, tt.expectedBootloader)
			}
		})
	}
}

// TestPreflightPhasePostCheck tests configuration validation after form submission
func TestPreflightPhasePostCheck(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		hostname    string
		timezone    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "All fields present - validation passes",
			username: "archuser",
			hostname: "my-arch",
			timezone: "UTC",
			wantErr:  false,
		},
		{
			name:        "Missing username - validation fails",
			username:    "",
			hostname:    "my-arch",
			timezone:    "UTC",
			wantErr:     true,
			errContains: "username is required",
		},
		{
			name:        "Missing hostname - validation fails",
			username:    "archuser",
			hostname:    "",
			timezone:    "UTC",
			wantErr:     true,
			errContains: "hostname is required",
		},
		{
			name:        "Missing timezone - validation fails",
			username:    "archuser",
			hostname:    "my-arch",
			timezone:    "",
			wantErr:     true,
			errContains: "timezone is required",
		},
		{
			name:        "All fields empty - validation fails on username",
			username:    "",
			hostname:    "",
			timezone:    "",
			wantErr:     true,
			errContains: "username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockCmd := mocks.NewMockCommandExecutor(ctrl)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			cfg.Username = tt.username
			cfg.Hostname = tt.hostname
			cfg.Timezone = tt.timezone
			cfg.ConfigPath = filepath.Join(tmpDir, "config") // Set config path to temp dir so Save() works

			phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)

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

// TestPreflightPhaseExecute tests the Execute method
func TestPreflightPhaseExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockCmd := mocks.NewMockCommandExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)

	progressChan := make(chan ProgressUpdate, 10)
	result := phase.Execute(progressChan)

	if !result.Success {
		t.Errorf("Execute() Success = %v, want true", result.Success)
	}

	if result.Message != "Preflight checks passed" {
		t.Errorf("Execute() Message = %q, want 'Preflight checks passed'", result.Message)
	}

	// Verify progress updates were sent
	updates := []ProgressUpdate{}
	for len(progressChan) > 0 {
		updates = append(updates, <-progressChan)
	}

	if len(updates) < 1 {
		t.Error("Execute() should send progress updates")
	}
}

// TestPreflightPhaseRollback tests the Rollback method
func TestPreflightPhaseRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockCmd := mocks.NewMockCommandExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)

	err = phase.Rollback()
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}
}

// TestPreflightPhaseCanSkip tests the CanSkip method
func TestPreflightPhaseCanSkip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockCmd := mocks.NewMockCommandExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewPreflightPhase(cfg, log, mockFS, mockCmd)

	if phase.CanSkip() {
		t.Error("CanSkip() = true, want false")
	}
}

// Helper functions are defined in helpers_test.go to avoid duplication
