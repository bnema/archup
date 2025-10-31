package packages

import "errors"

// KernelVariant represents available kernel choices
type KernelVariant int

const (
	// KernelStable is the stable Linux kernel
	KernelStable KernelVariant = iota

	// KernelZen is the linux-zen kernel
	KernelZen

	// KernelLTS is the linux-lts (long-term support) kernel
	KernelLTS

	// KernelHardened is the hardened kernel with security focus
	KernelHardened

	// KernelCachyOS is the CachyOS optimized kernel
	KernelCachyOS
)

// String returns the kernel package name
func (k KernelVariant) String() string {
	switch k {
	case KernelZen:
		return "linux-zen"
	case KernelLTS:
		return "linux-lts"
	case KernelHardened:
		return "linux-hardened"
	case KernelCachyOS:
		return "linux-cachyos"
	default:
		return "linux"
	}
}

// IsStable returns true if kernel is stable
func (k KernelVariant) IsStable() bool {
	return k == KernelStable
}

// IsLTS returns true if kernel is long-term support
func (k KernelVariant) IsLTS() bool {
	return k == KernelLTS
}

// IsHardened returns true if kernel has hardening
func (k KernelVariant) IsHardened() bool {
	return k == KernelHardened
}

// Kernel is an immutable value object for kernel selection
type Kernel struct {
	variant KernelVariant
}

// NewKernel creates a new Kernel value object
func NewKernel(variant KernelVariant) (*Kernel, error) {
	// All kernel variants are valid
	if variant < KernelStable || variant > KernelCachyOS {
		return nil, errors.New("invalid kernel variant")
	}

	return &Kernel{
		variant: variant,
	}, nil
}

// Variant returns the kernel variant
func (k *Kernel) Variant() KernelVariant {
	return k.variant
}

// PackageName returns the pacman package name
func (k *Kernel) PackageName() string {
	return k.variant.String()
}

// String returns human-readable representation
func (k *Kernel) String() string {
	return "Kernel(" + k.variant.String() + ")"
}

// Equals checks if two Kernel objects are equal
func (k *Kernel) Equals(other *Kernel) bool {
	if other == nil {
		return false
	}
	return k.variant == other.variant
}

// AvailableKernels returns all available kernel variants
func AvailableKernels() []KernelVariant {
	return []KernelVariant{
		KernelStable,
		KernelZen,
		KernelLTS,
		KernelHardened,
		KernelCachyOS,
	}
}
