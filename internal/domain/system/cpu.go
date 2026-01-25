package system

import "errors"

// CPUVendor represents the CPU vendor
type CPUVendor int

const (
	// CPUVendorUnknown represents an unknown CPU vendor
	CPUVendorUnknown CPUVendor = iota

	// CPUVendorIntel represents Intel CPUs
	CPUVendorIntel

	// CPUVendorAMD represents AMD CPUs
	CPUVendorAMD

	// CPUVendorARM represents ARM CPUs
	CPUVendorARM
)

// String returns human-readable vendor name
func (v CPUVendor) String() string {
	switch v {
	case CPUVendorIntel:
		return "Intel"
	case CPUVendorAMD:
		return "AMD"
	case CPUVendorARM:
		return "ARM"
	default:
		return "Unknown"
	}
}

// Microcode represents CPU microcode package
type Microcode int

const (
	// MicrocodeNone means no microcode updates
	MicrocodeNone Microcode = iota

	// MicrocodeIntel for Intel CPUs
	MicrocodeIntel

	// MicrocodeAMD for AMD CPUs
	MicrocodeAMD
)

// String returns microcode package name
func (m Microcode) String() string {
	switch m {
	case MicrocodeIntel:
		return "intel-ucode"
	case MicrocodeAMD:
		return "amd-ucode"
	default:
		return ""
	}
}

// CPUInfo is an immutable value object representing CPU information
// It includes vendor, model, and microcode availability
type CPUInfo struct {
	vendor    CPUVendor
	model     string // CPU model name
	microcode Microcode
}

// NewCPUInfo creates a new CPUInfo value object
func NewCPUInfo(vendor CPUVendor, model string, microcode Microcode) (*CPUInfo, error) {
	if model == "" {
		return nil, errors.New("CPU model cannot be empty")
	}

	if len(model) > 128 {
		return nil, errors.New("CPU model name too long")
	}

	// Validate microcode matches vendor
	if vendor == CPUVendorIntel && microcode == MicrocodeAMD {
		return nil, errors.New("AMD microcode cannot be used with Intel CPU")
	}

	if vendor == CPUVendorAMD && microcode == MicrocodeIntel {
		return nil, errors.New("intel microcode cannot be used with AMD CPU")
	}

	return &CPUInfo{
		vendor:    vendor,
		model:     model,
		microcode: microcode,
	}, nil
}

// Vendor returns the CPU vendor
func (c *CPUInfo) Vendor() CPUVendor {
	return c.vendor
}

// Model returns the CPU model name
func (c *CPUInfo) Model() string {
	return c.model
}

// Microcode returns the CPU microcode package
func (c *CPUInfo) Microcode() Microcode {
	return c.microcode
}

// RequiresMicrocode returns true if this CPU needs microcode updates
func (c *CPUInfo) RequiresMicrocode() bool {
	return c.microcode != MicrocodeNone
}

// String returns human-readable representation
func (c *CPUInfo) String() string {
	return "CPUInfo(vendor=" + c.vendor.String() +
		", model=" + c.model +
		", microcode=" + c.microcode.String() + ")"
}

// Equals checks if two CPUInfo objects are equal
func (c *CPUInfo) Equals(other *CPUInfo) bool {
	if other == nil {
		return false
	}
	return c.vendor == other.vendor &&
		c.model == other.model &&
		c.microcode == other.microcode
}
