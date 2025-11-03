package dto

// SystemInfo contains basic system information
type SystemInfo struct {
	Architecture   string
	IsUEFI         bool
	Distribution   string
	SecureBootEnabled bool
}

// CPUInfo contains CPU information
type CPUInfo struct {
	Model string
}

// PreflightResult is the result of preflight checks
type PreflightResult struct {
	SystemInfo     *SystemInfo
	CPUInfo        *CPUInfo
	ChecksPassed   bool
	Warnings       []string
	CriticalErrors []string
}
