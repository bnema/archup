package handlers

import (
	"context"
	"strings"

	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/system"
)

// GPUHandler handles GPU detection.
type GPUHandler struct {
	cmdExec ports.CommandExecutor
	logger  ports.Logger
}

// NewGPUHandler creates a new GPU handler.
func NewGPUHandler(cmdExec ports.CommandExecutor, logger ports.Logger) *GPUHandler {
	return &GPUHandler{
		cmdExec: cmdExec,
		logger:  logger,
	}
}

// Detect detects the system GPU and returns recommended configuration.
func (h *GPUHandler) Detect(ctx context.Context) (*system.GPU, error) {
	output, err := h.cmdExec.Execute(ctx, "lspci")
	if err != nil {
		h.logger.Warn("Failed to run lspci for GPU detection", "error", err)
		return system.NewGPU(system.GPUVendorUnknown, "", nil, nil), err
	}

	line := firstGPULine(string(output))
	if line == "" {
		h.logger.Warn("No GPU detected in lspci output")
		return system.NewGPU(system.GPUVendorUnknown, "", nil, nil), nil
	}

	vendor := detectGPUVendor(line)
	model := extractGPUModel(line)
	drivers := recommendedDrivers(vendor)

	h.logger.Info("Detected GPU", "vendor", vendor, "model", model)
	return system.NewGPU(vendor, model, drivers, nil), nil
}

func firstGPULine(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA compatible controller") ||
			strings.Contains(line, "3D controller") ||
			strings.Contains(line, "Display controller") {
			return line
		}
	}
	return ""
}

func detectGPUVendor(line string) system.GPUVendor {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "nvidia"):
		return system.GPUVendorNVIDIA
	case strings.Contains(lower, "advanced micro devices") || strings.Contains(lower, "amd") || strings.Contains(lower, "ati"):
		return system.GPUVendorAMD
	case strings.Contains(lower, "intel"):
		return system.GPUVendorIntel
	default:
		return system.GPUVendorUnknown
	}
}

func extractGPUModel(line string) string {
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) < 2 {
		return "Unknown"
	}
	return strings.TrimSpace(parts[1])
}

func recommendedDrivers(vendor system.GPUVendor) []string {
	switch vendor {
	case system.GPUVendorAMD:
		return []string{"mesa", "vulkan-radeon", "libva-mesa-driver", "mesa-vdpau"}
	case system.GPUVendorIntel:
		return []string{"mesa", "vulkan-intel", "intel-media-driver"}
	case system.GPUVendorNVIDIA:
		return []string{"nvidia-open", "nvidia-utils", "libva-nvidia-driver"}
	default:
		return []string{}
	}
}
