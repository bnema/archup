package system

import (
	"testing"
)

// TestDetectTimezoneSuccess tests successful timezone detection
func TestDetectTimezoneSuccess(t *testing.T) {
	// Skip this test as it requires network and a System instance with logger
	// A proper unit test would require refactoring DetectTimezone to accept a base URL
	t.Skip("Skipping timezone detection test - requires System instance and network access")
}

// TestValidateTimezone tests timezone validation
func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		want     bool
	}{
		{"Valid Europe timezone", "Europe/Paris", true},
		{"Valid US timezone", "America/New_York", true},
		{"Valid Asia timezone", "Asia/Tokyo", true},
		{"Invalid - no slash", "UTC", false},
		{"Invalid - empty string", "", false},
		{"Invalid - just slash", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTimezone(tt.timezone)
			if got != tt.want {
				t.Errorf("ValidateTimezone(%q) = %v, want %v", tt.timezone, got, tt.want)
			}
		})
	}
}
