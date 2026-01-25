package bootloader

import (
	"testing"
)

func TestNewBootloader_Valid(t *testing.T) {
	bootloader, err := NewBootloader(BootloaderTypeLimine, 0, "ArchUp")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bootloader.Type() != BootloaderTypeLimine {
		t.Errorf("expected Limine type")
	}

	if bootloader.Timeout() != 0 {
		t.Errorf("expected timeout 0")
	}

	if bootloader.Branding() != "ArchUp" {
		t.Errorf("expected branding ArchUp")
	}
}

func TestNewBootloader_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   int
		shouldErr bool
	}{
		{"negative timeout", -1, true},
		{"zero timeout", 0, false},
		{"valid timeout", 5, false},
		{"max timeout", 600, false},
		{"exceed max", 601, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBootloader(BootloaderTypeLimine, tt.timeout, "Test")
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestNewBootloader_InvalidBranding(t *testing.T) {
	tests := []struct {
		name      string
		branding  string
		shouldErr bool
	}{
		{"empty branding", "", true},
		{"valid branding", "MyBrand", false},
		{"too long", "a" + string(make([]byte, 100)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBootloader(BootloaderTypeLimine, 0, tt.branding)
			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, err)
			}
		})
	}
}

func TestBootloaderType_String(t *testing.T) {
	tests := []struct {
		btype BootloaderType
		want  string
	}{
		{BootloaderTypeLimine, "Limine"},
	}

	for _, tt := range tests {
		if got := tt.btype.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestBootloader_IsLimine(t *testing.T) {
	limine, _ := NewBootloader(BootloaderTypeLimine, 0, "Test")

	if !limine.IsLimine() {
		t.Error("expected Limine")
	}
}

func TestBootloader_Equals(t *testing.T) {
	b1, _ := NewBootloader(BootloaderTypeLimine, 0, "Test")
	b2, _ := NewBootloader(BootloaderTypeLimine, 0, "Test")

	if !b1.Equals(b2) {
		t.Error("expected equal bootloaders")
	}

	if b1.Equals(nil) {
		t.Error("expected nil to not be equal")
	}
}
