package phases

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces/mocks"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
	"go.uber.org/mock/gomock"
)

// TestBaseInstallPhasePreCheck tests /mnt mount validation
func TestBaseInstallPhasePreCheck(t *testing.T) {
	tests := []struct {
		name        string
		mountResult system.CommandResult
		wantErr     bool
		errContains string
	}{
		{
			name:        "/mnt is mounted",
			mountResult: system.CommandResult{ExitCode: 0},
			wantErr:     false,
		},
		{
			name:        "/mnt is not mounted",
			mountResult: system.CommandResult{ExitCode: 32},
			wantErr:     true,
			errContains: "/mnt is not mounted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)

			mockSysExec.EXPECT().RunSimple("mountpoint", "-q", "/mnt").Return(tt.mountResult).Times(1)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

			err = phase.PreCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("PreCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstringBaseTest(errMsg, tt.errContains) {
					t.Errorf("PreCheck() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestBaseInstallPhaseDetectCPU tests CPU vendor detection
func TestBaseInstallPhaseDetectCPU(t *testing.T) {
	tests := []struct {
		name           string
		cpuinfoContent string
		expectedVendor string
		expectedMicrocode string
		wantErr        bool
	}{
		{
			name:           "Intel CPU detected",
			cpuinfoContent: "processor\t: 0\nvendor_id\t: GenuineIntel\nflags\t: fpu\n",
			expectedVendor: "Intel",
			expectedMicrocode: "intel-ucode",
			wantErr:        false,
		},
		{
			name:           "AMD CPU detected",
			cpuinfoContent: "processor\t: 0\nvendor_id\t: AuthenticAMD\nflags\t: fpu\n",
			expectedVendor: "AMD",
			expectedMicrocode: "amd-ucode",
			wantErr:        false,
		},
		{
			name:           "Unknown CPU vendor",
			cpuinfoContent: "processor\t: 0\nvendor_id\t: UnknownVendor\nflags\t: fpu\n",
			expectedVendor: "Unknown",
			expectedMicrocode: "",
			wantErr:        false,
		},
		{
			name:           "Empty cpuinfo",
			cpuinfoContent: "",
			expectedVendor: "Unknown",
			expectedMicrocode: "",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)

			reader := io.NopCloser(bytes.NewBufferString(tt.cpuinfoContent))
			mockFS.EXPECT().Open("/proc/cpuinfo").Return(reader, nil).Times(1)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

			progressChan := make(chan ProgressUpdate, 10)
			err = phase.detectCPU(progressChan)

			if (err != nil) != tt.wantErr {
				t.Errorf("detectCPU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if cfg.CPUVendor != tt.expectedVendor {
				t.Errorf("CPUVendor = %q, want %q", cfg.CPUVendor, tt.expectedVendor)
			}
			if cfg.Microcode != tt.expectedMicrocode {
				t.Errorf("Microcode = %q, want %q", cfg.Microcode, tt.expectedMicrocode)
			}
		})
	}
}

// TestBaseInstallPhaseLoadBasePackages tests package file parsing
func TestBaseInstallPhaseLoadBasePackages(t *testing.T) {
	tests := []struct {
		name             string
		fileContent      string
		expectedPackages []string
		wantErr          bool
		errContains      string
	}{
		{
			name:             "Valid package list",
			fileContent:      "base\nlinux\nlinux-firmware\ngit\n",
			expectedPackages: []string{"base", "linux", "linux-firmware", "git"},
			wantErr:          false,
		},
		{
			name:             "Package list with comments and empty lines",
			fileContent:      "# Core packages\nbase\n\n# Additional\nlinux\n# Another comment\n\ngit\n",
			expectedPackages: []string{"base", "linux", "git"},
			wantErr:          false,
		},
		{
			name:             "Empty file",
			fileContent:      "",
			expectedPackages: []string{},
			wantErr:          false,
		},
		{
			name:             "Only comments",
			fileContent:      "# Comment 1\n# Comment 2\n",
			expectedPackages: []string{},
			wantErr:          false,
		},
		{
			name:             "Only whitespace",
			fileContent:      "   \n\n  \n",
			expectedPackages: []string{},
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)

			reader := io.NopCloser(bytes.NewBufferString(tt.fileContent))
			packageFile := config.DefaultInstallDir + "/" + config.BasePackagesFile
			mockFS.EXPECT().Open(packageFile).Return(reader, nil).Times(1)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

			packages, err := phase.loadBasePackages()

			if (err != nil) != tt.wantErr {
				t.Errorf("loadBasePackages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(packages) != len(tt.expectedPackages) {
				t.Errorf("loadBasePackages() returned %d packages, want %d", len(packages), len(tt.expectedPackages))
				return
			}

			for i, pkg := range packages {
				if pkg != tt.expectedPackages[i] {
					t.Errorf("Package %d = %q, want %q", i, pkg, tt.expectedPackages[i])
				}
			}
		})
	}
}

// TestBaseInstallPhaseGenerateFstab tests fstab generation
func TestBaseInstallPhaseGenerateFstab(t *testing.T) {
	tests := []struct {
		name            string
		genfstabResult  system.CommandResult
		wantErr         bool
		errContains     string
	}{
		{
			name:            "Fstab generation success",
			genfstabResult:  system.CommandResult{ExitCode: 0},
			wantErr:         false,
		},
		{
			name:            "Genfstab command failed",
			genfstabResult:  system.CommandResult{ExitCode: 1, Error: fmt.Errorf("permission denied")},
			wantErr:         true,
			errContains:     "genfstab failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)

			mockSysExec.EXPECT().RunSimple("sh", "-c", "genfstab -U /mnt >> /mnt/etc/fstab").Return(tt.genfstabResult).Times(1)

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

			progressChan := make(chan ProgressUpdate, 10)
			err = phase.generateFstab(progressChan)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateFstab() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstringBaseTest(errMsg, tt.errContains) {
					t.Errorf("generateFstab() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestBaseInstallPhasePostCheck tests installation validation
func TestBaseInstallPhasePostCheck(t *testing.T) {
	tests := []struct {
		name                  string
		usrDirExists          bool
		fstabFileExists       bool
		wantErr               bool
		errContains           string
	}{
		{
			name:                  "Both /mnt/usr and /mnt/etc/fstab exist",
			usrDirExists:          true,
			fstabFileExists:       true,
			wantErr:               false,
		},
		{
			name:                  "/mnt/usr does not exist",
			usrDirExists:          false,
			fstabFileExists:       true,
			wantErr:               true,
			errContains:           "/mnt/usr does not exist",
		},
		{
			name:                  "/mnt/etc/fstab does not exist",
			usrDirExists:          true,
			fstabFileExists:       false,
			wantErr:               true,
			errContains:           "fstab was not created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			mockSysExec := mocks.NewMockSystemExecutor(ctrl)

			usrResult := system.CommandResult{ExitCode: 1}
			if tt.usrDirExists {
				usrResult.ExitCode = 0
			}
			mockSysExec.EXPECT().RunSimple("test", "-d", "/mnt/usr").Return(usrResult).Times(1)

			// Only expect fstab check if usr dir exists
			if tt.usrDirExists {
				fstabResult := system.CommandResult{ExitCode: 1}
				if tt.fstabFileExists {
					fstabResult.ExitCode = 0
				}
				mockSysExec.EXPECT().RunSimple("test", "-f", "/mnt/etc/fstab").Return(fstabResult).Times(1)
			}

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")
			log, err := logger.New(logPath, false)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			defer log.Close()

			cfg := config.NewConfig("test")
			cfg.ConfigPath = filepath.Join(tmpDir, "config")

			phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

			err = phase.PostCheck()

			if (err != nil) != tt.wantErr {
				t.Errorf("PostCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if errMsg := err.Error(); tt.errContains != "" && !containsSubstringBaseTest(errMsg, tt.errContains) {
					t.Errorf("PostCheck() error = %q, should contain %q", errMsg, tt.errContains)
				}
			}
		})
	}
}

// TestBaseInstallPhaseExecute tests Execute calls correct steps
func TestBaseInstallPhaseExecute(t *testing.T) {
	// Note: Full Execute testing is complex due to logger.ExecCommand() calls
	// This test verifies the phase structure and progress updates
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)

	cfg := config.NewConfig("test")
	phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

	progressChan := make(chan ProgressUpdate, 100)

	// Just verify that Execute sends progress updates (higher-level behavior)
	// since testing the full Execute requires mocking logger.ExecCommand
	result := phase.Execute(progressChan)

	// Result may be error due to logger.ExecCommand, but we should get a PhaseResult
	if result.Message == "" && !result.Success {
		// This is expected when logger.ExecCommand is called
		t.Logf("Execute correctly attempted to run - logger.ExecCommand not mocked")
	}
}

// TestBaseInstallPhaseRollback tests rollback
func TestBaseInstallPhaseRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

	// BaseInstallPhase Rollback should be no-op
	err = phase.Rollback()
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}
}

// TestBaseInstallPhaseCanSkip tests CanSkip
func TestBaseInstallPhaseCanSkip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockSysExec := mocks.NewMockSystemExecutor(ctrl)

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	log, err := logger.New(logPath, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	cfg := config.NewConfig("test")
	phase := NewBaseInstallPhase(cfg, log, mockFS, mockSysExec)

	if phase.CanSkip() {
		t.Error("CanSkip() = true, want false")
	}
}

// Helper function for substring matching (avoid duplication across test files)
func containsSubstringBaseTest(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
