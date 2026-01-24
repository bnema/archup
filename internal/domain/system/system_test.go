package system

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

// SystemConfig Tests

func TestNewSystemConfig_Valid(t *testing.T) {
	config, err := NewSystemConfig("myhost", "UTC", "en_US.UTF-8", "us")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.Hostname() != "myhost" {
		t.Errorf("expected hostname 'myhost', got '%s'", config.Hostname())
	}

	if config.Timezone() != "UTC" {
		t.Errorf("expected timezone 'UTC', got '%s'", config.Timezone())
	}

	if config.Locale() != "en_US.UTF-8" {
		t.Errorf("expected locale 'en_US.UTF-8', got '%s'", config.Locale())
	}

	if config.Keymap() != "us" {
		t.Errorf("expected keymap 'us', got '%s'", config.Keymap())
	}
}

func TestNewSystemConfig_InvalidHostname(t *testing.T) {
	tests := []struct {
		name      string
		hostname  string
		shouldErr bool
	}{
		{"empty hostname", "", true},
		{"starts with hyphen", "-host", true},
		{"ends with hyphen", "host-", true},
		{"too long", "a" + string(make([]byte, 100)), true},
		{"valid hostname", "myhost", false},
		{"with hyphen", "my-host", false},
		{"with digits", "host123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSystemConfig(tt.hostname, "UTC", "en_US.UTF-8", "us")
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestNewSystemConfig_InvalidLocale(t *testing.T) {
	tests := []struct {
		name      string
		locale    string
		shouldErr bool
	}{
		{"valid locale", "en_US.UTF-8", false},
		{"empty locale allowed", "", false},
		{"C locale", "C", false},
		{"C.UTF-8", "C.UTF-8", false},
		{"too long", "a" + string(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSystemConfig("host", "UTC", tt.locale, "us")
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestNewSystemConfig_InvalidKeymap(t *testing.T) {
	tests := []struct {
		name      string
		keymap    string
		shouldErr bool
	}{
		{"valid simple keymap", "us", false},
		{"with variant", "de-nodeadkeys", false},
		{"empty keymap", "", true},
		{"with uppercase", "US", true},
		{"with special chars", "us!", true},
		{"too long", "a" + string(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSystemConfig("host", "UTC", "en_US.UTF-8", tt.keymap)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestSystemConfig_Equals(t *testing.T) {
	config1, _ := NewSystemConfig("host1", "UTC", "en_US.UTF-8", "us")
	config2, _ := NewSystemConfig("host1", "UTC", "en_US.UTF-8", "us")
	config3, _ := NewSystemConfig("host2", "UTC", "en_US.UTF-8", "us")

	if !config1.Equals(config2) {
		t.Error("expected equal configs")
	}

	if config1.Equals(config3) {
		t.Error("expected different configs")
	}

	if config1.Equals(nil) {
		t.Error("expected nil to not be equal")
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name      string
		hostname  string
		shouldErr bool
	}{
		{"valid", "myhost", false},
		{"with digits", "host123", false},
		{"with hyphen", "my-host", false},
		{"empty", "", true},
		{"starts with hyphen", "-host", true},
		{"ends with hyphen", "host-", true},
		{"contains underscore", "my_host", true},
		{"contains space", "my host", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		shouldErr bool
	}{
		{"UTC", "UTC", false},
		{"valid timezone", "America/New_York", false},
		{"empty allowed", "", false},
		{"contains ..", "UTC..", true},
		{"starts with /", "/UTC", true},
		{"too long", "a" + string(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.timezone)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateLocale(t *testing.T) {
	tests := []struct {
		name      string
		locale    string
		shouldErr bool
	}{
		{"valid UTF-8", "en_US.UTF-8", false},
		{"C locale", "C", false},
		{"empty allowed", "", false},
		{"ISO variant", "en_US.ISO8859-1", false},
		{"too long", "a" + string(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocale(tt.locale)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateKeymap(t *testing.T) {
	tests := []struct {
		name      string
		keymap    string
		shouldErr bool
	}{
		{"us", "us", false},
		{"de", "de", false},
		{"with variant", "de-nodeadkeys", false},
		{"empty", "", true},
		{"uppercase", "US", true},
		{"with special chars", "us!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeymap(tt.keymap)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

// CPUInfo Tests

func TestNewCPUInfo_Valid(t *testing.T) {
	cpuInfo, err := NewCPUInfo(CPUVendorIntel, "Intel(R) Core(TM) i7-9700K", MicrocodeIntel)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cpuInfo.Vendor() != CPUVendorIntel {
		t.Errorf("expected Intel vendor")
	}

	if !cpuInfo.RequiresMicrocode() {
		t.Error("expected to require microcode")
	}
}

func TestNewCPUInfo_InvalidModel(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		shouldErr bool
	}{
		{"valid", "Intel Core i7", false},
		{"empty", "", true},
		{"too long", "a" + string(make([]byte, 200)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCPUInfo(CPUVendorIntel, tt.model, MicrocodeIntel)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestNewCPUInfo_MicrocodeVendorMismatch(t *testing.T) {
	tests := []struct {
		name      string
		vendor    CPUVendor
		microcode Microcode
		shouldErr bool
	}{
		{"Intel with Intel ucode", CPUVendorIntel, MicrocodeIntel, false},
		{"Intel with AMD ucode", CPUVendorIntel, MicrocodeAMD, true},
		{"AMD with AMD ucode", CPUVendorAMD, MicrocodeAMD, false},
		{"AMD with Intel ucode", CPUVendorAMD, MicrocodeIntel, true},
		{"Intel with no ucode", CPUVendorIntel, MicrocodeNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCPUInfo(tt.vendor, "Model", tt.microcode)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestCPUInfo_Equals(t *testing.T) {
	cpu1, _ := NewCPUInfo(CPUVendorIntel, "i7", MicrocodeIntel)
	cpu2, _ := NewCPUInfo(CPUVendorIntel, "i7", MicrocodeIntel)
	cpu3, _ := NewCPUInfo(CPUVendorAMD, "Ryzen", MicrocodeAMD)

	if !cpu1.Equals(cpu2) {
		t.Error("expected equal CPUInfo")
	}

	if cpu1.Equals(cpu3) {
		t.Error("expected different CPUInfo")
	}

	if cpu1.Equals(nil) {
		t.Error("expected nil to not be equal")
	}
}

func TestCPUVendor_String(t *testing.T) {
	tests := []struct {
		vendor CPUVendor
		want   string
	}{
		{CPUVendorIntel, "Intel"},
		{CPUVendorAMD, "AMD"},
		{CPUVendorARM, "ARM"},
		{CPUVendorUnknown, "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.vendor.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestMicrocode_String(t *testing.T) {
	tests := []struct {
		microcode Microcode
		want      string
	}{
		{MicrocodeIntel, "intel-ucode"},
		{MicrocodeAMD, "amd-ucode"},
		{MicrocodeNone, ""},
	}

	for _, tt := range tests {
		if got := tt.microcode.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

// ValidationRules Tests

func TestSanitizeHostname(t *testing.T) {
	rules := NewSystemValidationRules()

	tests := []struct {
		input    string
		expected string
	}{
		{"MyHost", "myhost"},
		{"my-host", "my-host"},
		{"my_host", "my-host"},
		{"-myhost", "myhost"},
		{"myhost-", "myhost"},
		{"my--host", "my-host"},
		{"my__host", "my-host"},
		{"MyHost123", "myhost123"},
		{strings.Repeat("a", 100), strings.Repeat("a", 63)},
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(10, len(tt.input))], func(t *testing.T) {
			got := rules.SanitizeHostname(tt.input)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}

			// Verify result passes validation
			if err := ValidateHostname(got); err != nil {
				t.Errorf("sanitized hostname failed validation: %v", err)
			}
		})
	}
}

func TestValidateCombination(t *testing.T) {
	rules := NewSystemValidationRules()

	config, _ := NewSystemConfig("host", "UTC", "en_US.UTF-8", "us")
	cpu, _ := NewCPUInfo(CPUVendorIntel, "i7", MicrocodeIntel)

	err := rules.ValidateCombination(config, cpu)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test nil checks
	if err := rules.ValidateCombination(nil, cpu); err == nil {
		t.Error("expected error for nil config")
	}

	if err := rules.ValidateCombination(config, nil); err == nil {
		t.Error("expected error for nil cpu")
	}
}

func TestSuggestedLocaleForTimezone(t *testing.T) {
	rules := NewSystemValidationRules()

	tests := []struct {
		timezone string
		expected string
	}{
		{"UTC", "en_US.UTF-8"},
		{"America/New_York", "en_US.UTF-8"},
		{"Europe/London", "en_GB.UTF-8"},
		{"Europe/Paris", "fr_FR.UTF-8"},
		{"Unknown/Zone", ""},
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			got := rules.SuggestedLocaleForTimezone(tt.timezone)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestSuggestedKeymapForLocale(t *testing.T) {
	rules := NewSystemValidationRules()

	tests := []struct {
		locale   string
		expected string
	}{
		{"en_US.UTF-8", "us"},
		{"en_GB.UTF-8", "gb"},
		{"de_DE.UTF-8", "de"},
		{"fr_FR.UTF-8", "fr"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			got := rules.SuggestedKeymapForLocale(tt.locale)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// New validation function tests

func TestValidateArchLinux(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rules := NewSystemValidationRules()
	ctx := context.Background()

	tests := []struct {
		name       string
		fileExists bool
		shouldErr  bool
	}{
		{"arch-release exists", true, false},
		{"arch-release missing", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(ctrl)
			mockFS.EXPECT().Exists("/etc/arch-release").Return(tt.fileExists, nil)

			err := rules.ValidateArchLinux(ctx, mockFS)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateNotDerivative(t *testing.T) {
	rules := NewSystemValidationRules()
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(*mocks.MockFileSystem)
		shouldErr bool
		errMsg    string
	}{
		{
			name: "vanilla arch",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(false, nil)
			},
			shouldErr: false,
		},
		{
			name: "cachyos detected",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				// Map iteration order is undefined, so any may be checked first
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(true, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(false, nil).AnyTimes()
			},
			shouldErr: true,
			errMsg:    "CachyOS",
		},
		{
			name: "endeavouros detected",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/eos-release").Return(true, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(false, nil).AnyTimes()
			},
			shouldErr: true,
			errMsg:    "EndeavourOS",
		},
		{
			name: "garuda detected",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(true, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(false, nil).AnyTimes()
			},
			shouldErr: true,
			errMsg:    "Garuda",
		},
		{
			name: "manjaro detected",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil).AnyTimes()
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(true, nil).AnyTimes()
			},
			shouldErr: true,
			errMsg:    "Manjaro",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFS := mocks.NewMockFileSystem(ctrl)
			tt.setupMock(mockFS)

			err := rules.ValidateNotDerivative(ctx, mockFS)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
			if tt.shouldErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error to contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestValidateArchitecture(t *testing.T) {
	rules := NewSystemValidationRules()

	tests := []struct {
		name      string
		arch      string
		shouldErr bool
	}{
		{"x86_64", "x86_64", false},
		{"empty", "", true},
		{"aarch64", "aarch64", true},
		{"i686", "i686", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rules.ValidateArchitecture(tt.arch)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestValidateUEFIBoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rules := NewSystemValidationRules()
	ctx := context.Background()

	tests := []struct {
		name          string
		efivarsExists bool
		shouldErr     bool
	}{
		{"efivars exists", true, false},
		{"efivars missing", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(ctrl)
			mockFS.EXPECT().Exists("/sys/firmware/efi/efivars").Return(tt.efivarsExists, nil)

			err := rules.ValidateUEFIBoot(ctx, mockFS)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestDetectSecureBoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rules := NewSystemValidationRules()
	ctx := context.Background()

	tests := []struct {
		name     string
		output   []byte
		cmdErr   error
		expected bool
	}{
		{
			name:     "secure boot enabled",
			output:   []byte("System:\n  Secure Boot: enabled\n"),
			expected: true,
		},
		{
			name:     "secure boot disabled",
			output:   []byte("System:\n  Secure Boot: disabled\n"),
			expected: false,
		},
		{
			name:     "bootctl fails",
			cmdErr:   errors.New("bootctl not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := mocks.NewMockCommandExecutor(ctrl)
			if tt.cmdErr != nil {
				mockExec.EXPECT().Execute(ctx, "bootctl", "status").Return(nil, tt.cmdErr)
			} else {
				mockExec.EXPECT().Execute(ctx, "bootctl", "status").Return(tt.output, nil)
			}

			enabled, err := rules.DetectSecureBoot(ctx, mockExec)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if enabled != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, enabled)
			}
		})
	}
}

func TestValidateSecureBootDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rules := NewSystemValidationRules()
	ctx := context.Background()

	tests := []struct {
		name      string
		output    []byte
		shouldErr bool
	}{
		{
			name:      "secure boot disabled",
			output:    []byte("System:\n  Secure Boot: disabled\n"),
			shouldErr: false,
		},
		{
			name:      "secure boot enabled",
			output:    []byte("System:\n  Secure Boot: enabled\n"),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := mocks.NewMockCommandExecutor(ctrl)
			mockExec.EXPECT().Execute(ctx, "bootctl", "status").Return(tt.output, nil)

			err := rules.ValidateSecureBootDisabled(ctx, mockExec)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

func TestDistributionType_String(t *testing.T) {
	tests := []struct {
		distro DistributionType
		want   string
	}{
		{DistributionArch, "Arch Linux"},
		{DistributionCachyOS, "CachyOS"},
		{DistributionEndeavourOS, "EndeavourOS"},
		{DistributionGaruda, "Garuda Linux"},
		{DistributionManjaro, "Manjaro Linux"},
		{DistributionUnknown, "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.distro.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestDetectDistribution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(*mocks.MockFileSystem)
		expected  DistributionType
		shouldErr bool
	}{
		{
			name: "vanilla arch",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(true, nil)
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(false, nil)
			},
			expected: DistributionArch,
		},
		{
			name: "cachyos",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(true, nil)
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(true, nil)
			},
			expected: DistributionCachyOS,
		},
		{
			name: "endeavouros",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(true, nil)
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/eos-release").Return(true, nil)
			},
			expected: DistributionEndeavourOS,
		},
		{
			name: "garuda",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(true, nil)
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(true, nil)
			},
			expected: DistributionGaruda,
		},
		{
			name: "manjaro",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(true, nil)
				mockFS.EXPECT().Exists("/etc/cachyos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/eos-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/garuda-release").Return(false, nil)
				mockFS.EXPECT().Exists("/etc/manjaro-release").Return(true, nil)
			},
			expected: DistributionManjaro,
		},
		{
			name: "not arch",
			setupMock: func(mockFS *mocks.MockFileSystem) {
				mockFS.EXPECT().Exists("/etc/arch-release").Return(false, nil)
			},
			expected:  DistributionUnknown,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(ctrl)
			tt.setupMock(mockFS)

			distro, err := DetectDistribution(ctx, mockFS)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
			if distro != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, distro)
			}
		})
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
