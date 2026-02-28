package packages

import (
	"testing"
)

// Kernel Tests

func TestNewKernel_Valid(t *testing.T) {
	tests := []KernelVariant{
		KernelStable,
		KernelZen,
		KernelLTS,
		KernelHardened,
		KernelCachyOS,
	}

	for _, variant := range tests {
		t.Run(variant.String(), func(t *testing.T) {
			kernel, err := NewKernel(variant)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if kernel.Variant() != variant {
				t.Errorf("expected variant %v", variant)
			}
		})
	}
}

func TestKernelVariant_String(t *testing.T) {
	tests := []struct {
		variant KernelVariant
		want    string
	}{
		{KernelStable, "linux"},
		{KernelZen, "linux-zen"},
		{KernelLTS, "linux-lts"},
		{KernelHardened, "linux-hardened"},
		{KernelCachyOS, "linux-cachyos"},
	}

	for _, tt := range tests {
		if got := tt.variant.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestKernel_PackageName(t *testing.T) {
	kernel, _ := NewKernel(KernelZen)

	if kernel.PackageName() != "linux-zen" {
		t.Errorf("expected linux-zen package name")
	}
}

func TestKernel_Equals(t *testing.T) {
	k1, _ := NewKernel(KernelZen)
	k2, _ := NewKernel(KernelZen)
	k3, _ := NewKernel(KernelLTS)

	if !k1.Equals(k2) {
		t.Error("expected equal kernels")
	}

	if k1.Equals(k3) {
		t.Error("expected different kernels")
	}

	if k1.Equals(nil) {
		t.Error("expected nil to not be equal")
	}
}

func TestKernelVariant_IsLTS(t *testing.T) {
	tests := []struct {
		variant KernelVariant
		isLTS   bool
	}{
		{KernelStable, false},
		{KernelZen, false},
		{KernelLTS, true},
		{KernelHardened, false},
		{KernelCachyOS, false},
	}

	for _, tt := range tests {
		if got := tt.variant.IsLTS(); got != tt.isLTS {
			t.Errorf("%s: expected IsLTS=%v, got %v", tt.variant, tt.isLTS, got)
		}
	}
}

func TestKernelVariant_IsHardened(t *testing.T) {
	tests := []struct {
		variant  KernelVariant
		hardened bool
	}{
		{KernelStable, false},
		{KernelZen, false},
		{KernelLTS, false},
		{KernelHardened, true},
		{KernelCachyOS, false},
	}

	for _, tt := range tests {
		if got := tt.variant.IsHardened(); got != tt.hardened {
			t.Errorf("%s: expected IsHardened=%v, got %v", tt.variant, tt.hardened, got)
		}
	}
}

func TestAvailableKernels(t *testing.T) {
	kernels := AvailableKernels()

	if len(kernels) != 5 {
		t.Errorf("expected 5 kernels, got %d", len(kernels))
	}

	expected := []KernelVariant{
		KernelStable,
		KernelZen,
		KernelLTS,
		KernelHardened,
		KernelCachyOS,
	}

	for i, exp := range expected {
		if kernels[i] != exp {
			t.Errorf("expected kernel %v at index %d, got %v", exp, i, kernels[i])
		}
	}
}

// Repository Tests

func TestNewRepository_Valid(t *testing.T) {
	repo, err := NewRepository(true, AURHelperParu)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repo.EnableMultilib() {
		t.Error("expected multilib enabled")
	}

	if repo.AURHelper() != AURHelperParu {
		t.Error("expected paru AUR helper")
	}
}

func TestNewRepository_AllCombinations(t *testing.T) {
	tests := []struct {
		name     string
		multilib bool
		helper   AURHelper
	}{
		{"minimal", false, AURHelperParu},
		{"multilib", true, AURHelperParu},
		{"with yay", true, AURHelperYay},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewRepository(tt.multilib, tt.helper)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if repo.EnableMultilib() != tt.multilib {
				t.Errorf("expected multilib=%v", tt.multilib)
			}

			if repo.AURHelper() != tt.helper {
				t.Errorf("expected helper %v", tt.helper)
			}
		})
	}
}

func TestAURHelper_String(t *testing.T) {
	tests := []struct {
		helper AURHelper
		want   string
	}{
		{AURHelperParu, "paru"},
		{AURHelperYay, "yay"},
	}

	for _, tt := range tests {
		if got := tt.helper.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestRepository_Equals(t *testing.T) {
	r1, _ := NewRepository(true, AURHelperParu)
	r2, _ := NewRepository(true, AURHelperParu)
	r3, _ := NewRepository(true, AURHelperYay)

	if !r1.Equals(r2) {
		t.Error("expected equal repositories")
	}

	if r1.Equals(r3) {
		t.Error("expected different repositories")
	}

	if r1.Equals(nil) {
		t.Error("expected nil to not be equal")
	}
}
