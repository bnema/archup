package system

import (
	"strings"
	"testing"
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

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
