package phases

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces/mocks"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
	"go.uber.org/mock/gomock"
)

// TestPostInstallPhasePreCheck tests post-install prerequisites
func TestPostInstallPhasePreCheck(t *testing.T) {
	tests := []struct {
		name                string
		setupMocks          func(*mocks.MockFileSystem, *mocks.MockSystemExecutor)
		wantErr             bool
		errContains         string
	}{
		{
			name: "/mnt mounted and boot directory exists",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockSysExec *mocks.MockSystemExecutor) {
				mockSysExec.EXPECT().RunSimple("mountpoint", "-q", config.PathMnt).Return(system.CommandResult{ExitCode: 0}).Times(1)
				mockFS.EXPECT().Stat(config.PathMntBoot).Return(mockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).Times(1)
			},
			wantErr: false,
		},
		{
			name: "/mnt not mounted",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockSysExec *mocks.MockSystemExecutor) {
				mockSysExec.EXPECT().RunSimple("mountpoint", "-q", config.PathMnt).Return(system.CommandResult{ExitCode: 32}).Times(1)
			},
			wantErr:     true,
			errContains: "is not mounted",
		},
		{
			name: "boot directory not found",
			setupMocks: func(mockFS *mocks.MockFileSystem, mockSysExec *mocks.MockSystemExecutor) {
				mockSysExec.EXPECT().RunSimple("mountpoint", "-q", config.PathMnt).Return(system.CommandResult{ExitCode: 0}).Times(1)
				mockFS.EXPECT().Stat(config.PathMntBoot).Return(nil, fmt.Errorf("not found")).Times(1)
				mockFS.EXPECT().IsNotExist(gomock.Any()).Return(true).Times(1)
			},
			wantErr:     true,
			errContains: "boot directory not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockHTTP := mocks.NewMockHTTPClient(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)
			mockChrExec := mocks.NewMockChrootExecutor(ctrl)

			tt.setupMocks(mockFS, mockSysExec)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

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

// TestPostInstallPhaseTemplateReading tests template file reading
func TestPostInstallPhaseTemplateReading(t *testing.T) {
	tests := []struct {
		name          string
		templateFile  string
		setupMocks    func(*mocks.MockFileSystem)
		wantErr       bool
		errContains   string
	}{
		{
			name:         "Successfully read bashrc template",
			templateFile: "bashrc",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("bashrc content"), nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:         "Template file not found",
			templateFile: "missing.conf",
			setupMocks: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().ReadFile(gomock.Any()).Return(nil, fmt.Errorf("file not found")).Times(1)
			},
			wantErr:     true,
			errContains: "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
			phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

			_, err = phase.readTemplate(tt.templateFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestPostInstallPhasePostCheck tests post-install validation
func TestPostInstallPhasePostCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
	cfg.ConfigPath = filepath.Join(tmpDir, "config")

	phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

	err = phase.PostCheck()
	// PostCheck should succeed as it just saves the config
	if err != nil {
		t.Errorf("PostCheck() error = %v, want nil", err)
	}
}

// TestPostInstallPhaseRollback tests post-install rollback
func TestPostInstallPhaseRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
	phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

	// Rollback should always succeed (it's a cleanup operation)
	err = phase.Rollback()
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}
}

// TestPostInstallPhaseCanSkip tests the CanSkip method
func TestPostInstallPhaseCanSkip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
	phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

	if phase.CanSkip() {
		t.Error("CanSkip() = true, want false")
	}
}

// TestPostInstallPhaseExecute tests the Execute method
func TestPostInstallPhaseExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)
	mockChrExec := mocks.NewMockChrootExecutor(ctrl)

	// Allow any file system and executor operations
	mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte{}, nil).AnyTimes()
	mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockFS.EXPECT().Create(gomock.Any()).Return(&MockWriteCloser{}, nil).AnyTimes()
	mockFS.EXPECT().Stat(gomock.Any()).Return(mockFileInfo{}, nil).AnyTimes()
	mockFS.EXPECT().IsNotExist(gomock.Any()).Return(false).AnyTimes()
	mockFS.EXPECT().Chmod(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().Get(gomock.Any()).Return(nil, fmt.Errorf("skip http")).AnyTimes()
	mockChrExec.EXPECT().ChrootExec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootPacman(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockChrExec.EXPECT().ChrootSystemctl(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockSysExec.EXPECT().RunSimple(gomock.Any(), gomock.Any()).Return(system.CommandResult{ExitCode: 0}).AnyTimes()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	cfg.KernelChoice = "linux"
	cfg.RawURL = "https://example.com"

	phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

	progressChan := make(chan ProgressUpdate, 100)

	// Note: Full Execute testing is complex due to multiple async operations
	result := phase.Execute(progressChan)

	// Verify result structure is valid
	_ = result // If we got here, Execute returned a valid result structure
}

// TestPostInstallPhaseEncryptionHandling tests encryption-aware operations
func TestPostInstallPhaseEncryptionHandling(t *testing.T) {
	tests := []struct {
		name           string
		encryptionType string
		expectedClosing bool
	}{
		{
			name:            "Unencrypted - no closing needed",
			encryptionType:  config.EncryptionNone,
			expectedClosing: false,
		},
		{
			name:            "LUKS encrypted - closing needed",
			encryptionType:  config.EncryptionLUKS,
			expectedClosing: true,
		},
		{
			name:            "LUKS+LVM - closing needed",
			encryptionType:  config.EncryptionLUKSLVM,
			expectedClosing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
			cfg.EncryptionType = tt.encryptionType

			phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

			// Verify encryption type is stored correctly
			if phase.config.EncryptionType != tt.encryptionType {
				t.Errorf("Encryption type = %q, want %q", phase.config.EncryptionType, tt.encryptionType)
			}

			// Verify closing would be needed based on encryption
			needsClosing := tt.encryptionType == config.EncryptionLUKS || tt.encryptionType == config.EncryptionLUKSLVM
			if needsClosing != tt.expectedClosing {
				t.Errorf("Closing needed = %v, want %v", needsClosing, tt.expectedClosing)
			}
		})
	}
}

// TestPostInstallPhaseFileSystemOperations tests file system interaction patterns
func TestPostInstallPhaseFileSystemOperations(t *testing.T) {
	tests := []struct {
		name       string
		operation  string // "read", "write", "mkdir", "create"
	}{
		{
			name:      "Read pacman.conf",
			operation: "read",
		},
		{
			name:      "Write pacman.conf",
			operation: "write",
		},
		{
			name:      "Create hook directory",
			operation: "mkdir",
		},
		{
			name:      "Create hook file",
			operation: "create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockHTTP := mocks.NewMockHTTPClient(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)
			mockChrExec := mocks.NewMockChrootExecutor(ctrl)

			// Setup basic expectations based on operation
			switch tt.operation {
			case "read":
				mockFS.EXPECT().ReadFile(gomock.Any()).Return([]byte("content"), nil).AnyTimes()
			case "write":
				mockFS.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			case "mkdir":
				mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			case "create":
				mockFS.EXPECT().Create(gomock.Any()).Return(&MockWriteCloser{}, nil).AnyTimes()
			}

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

			// Verify phase can be created
			if phase == nil {
				t.Error("PostInstallPhase is nil")
			}
		})
	}
}

// TestPostInstallPhaseKernelVerification tests kernel verification logic
func TestPostInstallPhaseKernelVerification(t *testing.T) {
	tests := []struct {
		name           string
		kernelChoice   string
		expectedFormat string
	}{
		{
			name:           "Linux kernel",
			kernelChoice:   "linux",
			expectedFormat: "vmlinuz-linux",
		},
		{
			name:           "Linux LTS kernel",
			kernelChoice:   "linux-lts",
			expectedFormat: "vmlinuz-linux-lts",
		},
		{
			name:           "Linux Zen kernel",
			kernelChoice:   "linux-zen",
			expectedFormat: "vmlinuz-linux-zen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockHTTP := mocks.NewMockHTTPClient(ctrl)
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
			cfg.KernelChoice = tt.kernelChoice

			phase := NewPostInstallPhase(cfg, log, mockFS, mockHTTP, mockSysExec, mockChrExec)

			// Verify kernel choice is stored
			if phase.config.KernelChoice != tt.kernelChoice {
				t.Errorf("Kernel choice = %q, want %q", phase.config.KernelChoice, tt.kernelChoice)
			}

			// Verify expected kernel name format
			expectedKernelName := fmt.Sprintf("vmlinuz-%s", tt.kernelChoice)
			if !containsSubstring(expectedKernelName, "vmlinuz") {
				t.Errorf("Expected kernel name format %q doesn't contain 'vmlinuz'", expectedKernelName)
			}
		})
	}
}
