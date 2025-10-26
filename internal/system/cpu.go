package system

import (
	"fmt"
	"os"
	"strings"
)

// CPUVendor represents CPU manufacturer
type CPUVendor string

const (
	CPUVendorIntel   CPUVendor = "Intel"
	CPUVendorAMD     CPUVendor = "AMD"
	CPUVendorUnknown CPUVendor = "Unknown"
)

// AMDPStateMode represents AMD P-State driver mode
type AMDPStateMode string

const (
	AMDPStateModeActive  AMDPStateMode = "active"
	AMDPStateModeGuided  AMDPStateMode = "guided"
	AMDPStateModePassive AMDPStateMode = "passive"
)

// CPUInfo holds detected CPU information
type CPUInfo struct {
	Vendor         CPUVendor
	Microcode      string
	AMDPStateModes []AMDPStateMode
}

// DetectCPUVendor reads /proc/cpuinfo and returns the CPU vendor
func DetectCPUVendor() (CPUVendor, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	switch {
	case err != nil:
		return CPUVendorUnknown, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "vendor_id"):
			parts := strings.Split(line, ":")
			switch {
			case len(parts) < 2:
				continue
			}

			vendor := strings.TrimSpace(parts[1])
			switch vendor {
			case "GenuineIntel":
				return CPUVendorIntel, nil
			case "AuthenticAMD":
				return CPUVendorAMD, nil
			default:
				return CPUVendorUnknown, nil
			}
		}
	}

	return CPUVendorUnknown, nil
}

// DetectMicrocode returns the appropriate microcode package based on CPU vendor
func DetectMicrocode(vendor CPUVendor) string {
	switch vendor {
	case CPUVendorIntel:
		return "intel-ucode"
	case CPUVendorAMD:
		return "amd-ucode"
	default:
		return ""
	}
}

// CheckCPPCSupport checks if CPU supports CPPC (required for AMD P-State active mode)
func CheckCPPCSupport() bool {
	// Check /proc/cpuinfo for cppc flag
	data, err := os.ReadFile("/proc/cpuinfo")
	switch {
	case err != nil:
		return false
	}

	switch {
	case strings.Contains(string(data), "cppc"):
		return true
	}

	// Check if amd_pstate directory exists
	_, err = os.Stat("/sys/devices/system/cpu/cpu0/cpufreq/amd_pstate")
	switch {
	case err == nil:
		return true
	}

	return false
}

// DetectAMDPStateModes detects available AMD P-State modes
func DetectAMDPStateModes() []AMDPStateMode {
	var modes []AMDPStateMode

	// Try to read available modes from sysfs
	data, err := os.ReadFile("/sys/devices/system/cpu/amd_pstate/status")
	switch {
	case err == nil:
		// Parse available modes from status file
		content := string(data)
		switch {
		case strings.Contains(content, "active"):
			modes = append(modes, AMDPStateModeActive)
		}
		switch {
		case strings.Contains(content, "guided"):
			modes = append(modes, AMDPStateModeGuided)
		}
		switch {
		case strings.Contains(content, "passive"):
			modes = append(modes, AMDPStateModePassive)
		}

		// If we found modes, return them
		switch {
		case len(modes) > 0:
			return modes
		}
	}

	// Fallback: detect based on CPPC support
	switch CheckCPPCSupport() {
	case true:
		// Modern AMD Ryzen CPUs support all three modes
		return []AMDPStateMode{
			AMDPStateModeActive,
			AMDPStateModeGuided,
			AMDPStateModePassive,
		}
	default:
		// Older AMD CPUs only support passive mode
		return []AMDPStateMode{AMDPStateModePassive}
	}
}

// GetAMDPStateModeDescription returns human-readable description for P-State mode
func GetAMDPStateModeDescription(mode AMDPStateMode) string {
	switch mode {
	case AMDPStateModeActive:
		return "active - Best performance (recommended for desktop/gaming)"
	case AMDPStateModeGuided:
		return "guided - Balanced performance and efficiency (recommended for laptops)"
	case AMDPStateModePassive:
		return "passive - Maximum compatibility (older CPUs)"
	default:
		return string(mode)
	}
}

// DetectCPUInfo detects all CPU information in one call (DRY)
func DetectCPUInfo() (*CPUInfo, error) {
	vendor, err := DetectCPUVendor()
	switch {
	case err != nil:
		return nil, err
	}

	info := &CPUInfo{
		Vendor:    vendor,
		Microcode: DetectMicrocode(vendor),
	}

	// Detect AMD P-State modes only for AMD CPUs
	switch vendor {
	case CPUVendorAMD:
		info.AMDPStateModes = DetectAMDPStateModes()
	}

	return info, nil
}
