package system

import (
	"fmt"
	"os"
	"strconv"
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
	AMDPStateModeNone    AMDPStateMode = ""
)

// AMDZenGen represents AMD Zen generation
type AMDZenGen struct {
	Generation string // "1", "1+", "2", "3", "3+", "4", "5", "unknown"
	Label      string // "Zen 1", "Zen+", "Zen 2", etc.
}

// CPUInfo holds detected CPU information
type CPUInfo struct {
	Vendor         CPUVendor
	ModelName      string        // Full CPU model name
	Microcode      string
	AMDZenGen      *AMDZenGen    // AMD Zen generation (nil for non-AMD)
	AMDPStateModes []AMDPStateMode
	RecommendedPStateMode AMDPStateMode // Recommended mode for this CPU
}

// DetectCPUVendor reads /proc/cpuinfo and returns the CPU vendor
func DetectCPUVendor() (CPUVendor, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return CPUVendorUnknown, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "vendor_id"):
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
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
	if err != nil {
		return false
	}

	switch {
	case strings.Contains(string(data), "cppc"):
		return true
	}

	// Check if amd_pstate directory exists
	_, err = os.Stat("/sys/devices/system/cpu/cpu0/cpufreq/amd_pstate")
	if err == nil {
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
		if len(modes) > 0 {
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

// DetectAMDZenGeneration detects AMD Zen generation based on CPU family and model
// Reference: https://en.wikichip.org/wiki/amd/cpuid
func DetectAMDZenGeneration() (*AMDZenGen, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	var cpuFamily, cpuModel int
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "cpu family") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val := strings.TrimSpace(parts[1])
				cpuFamily, _ = strconv.Atoi(val)
			}
		}
		if strings.HasPrefix(line, "model") && !strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val := strings.TrimSpace(parts[1])
				cpuModel, _ = strconv.Atoi(val)
			}
		}

		// Break early if we have both values
		if cpuFamily > 0 && cpuModel >= 0 {
			break
		}
	}

	zenGen := &AMDZenGen{
		Generation: "unknown",
		Label:      "Unknown",
	}

	// Detect Zen generation based on CPU family and model
	switch cpuFamily {
	case 23: // Family 17h
		if cpuModel >= 0 && cpuModel <= 15 {
			zenGen.Generation = "1"
			zenGen.Label = "Zen 1"
		} else if cpuModel >= 16 && cpuModel <= 31 {
			zenGen.Generation = "1+"
			zenGen.Label = "Zen+"
		} else if cpuModel >= 48 && cpuModel <= 127 {
			zenGen.Generation = "2"
			zenGen.Label = "Zen 2"
		} else {
			zenGen.Generation = "2"
			zenGen.Label = "Zen 2 (assumed)"
		}
	case 25: // Family 19h
		if cpuModel >= 0 && cpuModel <= 15 {
			zenGen.Generation = "3"
			zenGen.Label = "Zen 3"
		} else if cpuModel >= 32 && cpuModel <= 47 {
			zenGen.Generation = "3"
			zenGen.Label = "Zen 3"
		} else if cpuModel >= 80 && cpuModel <= 95 {
			zenGen.Generation = "3+"
			zenGen.Label = "Zen 3+"
		} else if cpuModel >= 96 && cpuModel <= 175 {
			zenGen.Generation = "4"
			zenGen.Label = "Zen 4"
		} else {
			zenGen.Generation = "3"
			zenGen.Label = "Zen 3 (assumed)"
		}
	case 26: // Family 1Ah
		zenGen.Generation = "5"
		zenGen.Label = "Zen 5"
	}

	return zenGen, nil
}

// GetRecommendedPStateMode returns the best recommended mode for each Zen generation
func GetRecommendedPStateMode(zenGen *AMDZenGen) AMDPStateMode {
	switch zenGen.Generation {
	case "1", "1+":
		return AMDPStateModeNone // Use default acpi-cpufreq
	case "2":
		return AMDPStateModePassive // Conservative choice for Zen 2
	case "3", "3+":
		return AMDPStateModeActive // Best for Zen 3/3+
	case "4", "5":
		return AMDPStateModeActive // Best for Zen 4/5
	default:
		return AMDPStateModePassive
	}
}

// GetAvailablePStateModes returns available P-State modes for a given Zen generation
func GetAvailablePStateModes(zenGen *AMDZenGen) []AMDPStateMode {
	switch zenGen.Generation {
	case "1", "1+":
		// Zen 1 and Zen+: No CPPC support, use acpi-cpufreq
		return []AMDPStateMode{}
	case "2":
		// Zen 2: Basic CPPC support, passive and active modes
		return []AMDPStateMode{AMDPStateModePassive, AMDPStateModeActive}
	case "3", "3+", "4", "5":
		// Zen 3+: Full CPPC support for all modes
		return []AMDPStateMode{AMDPStateModeActive, AMDPStateModeGuided, AMDPStateModePassive}
	default:
		return []AMDPStateMode{AMDPStateModePassive}
	}
}

// GetPStateModeDescription returns human-readable description for P-State mode
func GetPStateModeDescription(mode AMDPStateMode) string {
	switch mode {
	case AMDPStateModeActive:
		return "CPPC autonomous with EPP hints (best performance, desktop/gaming)"
	case AMDPStateModeGuided:
		return "CPPC guided autonomous (balanced, laptop/hybrid)"
	case AMDPStateModePassive:
		return "CPPC non-autonomous (compatibility, manual governor control)"
	default:
		return string(mode)
	}
}

// GetAMDPStateKernelParams returns kernel parameters needed for AMD P-State mode
func GetAMDPStateKernelParams(zenGen *AMDZenGen, mode AMDPStateMode) string {
	if mode == AMDPStateModeNone || mode == "" {
		return ""
	}

	// Some Zen 2 CPUs need shared_mem parameter for active mode
	if zenGen.Generation == "2" && mode == AMDPStateModeActive {
		return fmt.Sprintf("amd_pstate=%s amd_pstate.shared_mem=1", mode)
	}

	return fmt.Sprintf("amd_pstate=%s", mode)
}

// GetCPUModelName extracts the CPU model name from /proc/cpuinfo
func GetCPUModelName() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// DetectCPUInfo detects all CPU information in one call (DRY)
func DetectCPUInfo() (*CPUInfo, error) {
	vendor, err := DetectCPUVendor()
	if err != nil {
		return nil, err
	}

	info := &CPUInfo{
		Vendor:    vendor,
		ModelName: GetCPUModelName(),
		Microcode: DetectMicrocode(vendor),
	}

	// Detect AMD-specific features only for AMD CPUs
	if vendor == CPUVendorAMD {
		zenGen, err := DetectAMDZenGeneration()
		if err == nil {
			info.AMDZenGen = zenGen
			info.AMDPStateModes = GetAvailablePStateModes(zenGen)
			info.RecommendedPStateMode = GetRecommendedPStateMode(zenGen)
		}
	}

	return info, nil
}
