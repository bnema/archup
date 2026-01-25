package system

import "strings"

// GPUVendor represents the GPU vendor.
type GPUVendor string

const (
	GPUVendorAMD     GPUVendor = "amd"
	GPUVendorIntel   GPUVendor = "intel"
	GPUVendorNVIDIA  GPUVendor = "nvidia"
	GPUVendorUnknown GPUVendor = "unknown"
)

// String returns human-readable vendor name.
func (v GPUVendor) String() string {
	switch v {
	case GPUVendorAMD:
		return "AMD"
	case GPUVendorIntel:
		return "Intel"
	case GPUVendorNVIDIA:
		return "NVIDIA"
	default:
		return "Unknown"
	}
}

// GPU represents a detected GPU and its recommended configuration.
type GPU struct {
	vendor  GPUVendor
	model   string
	drivers []string
	envVars map[string]string
}

// NewGPU creates a new GPU value object.
func NewGPU(vendor GPUVendor, model string, drivers []string, envVars map[string]string) *GPU {
	if strings.TrimSpace(model) == "" {
		model = "Unknown"
	}

	if drivers == nil {
		drivers = []string{}
	}

	if envVars == nil {
		envVars = map[string]string{}
	}

	return &GPU{
		vendor:  vendor,
		model:   model,
		drivers: drivers,
		envVars: envVars,
	}
}

// Vendor returns the GPU vendor.
func (g *GPU) Vendor() GPUVendor {
	return g.vendor
}

// Model returns the GPU model name.
func (g *GPU) Model() string {
	return g.model
}

// Drivers returns recommended GPU driver packages.
func (g *GPU) Drivers() []string {
	return append([]string{}, g.drivers...)
}

// EnvVars returns recommended environment variables for compositors.
func (g *GPU) EnvVars() map[string]string {
	copyVars := make(map[string]string, len(g.envVars))
	for key, value := range g.envVars {
		copyVars[key] = value
	}
	return copyVars
}
