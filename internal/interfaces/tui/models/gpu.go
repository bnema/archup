package models

import "github.com/bnema/archup/internal/domain/system"

// GPUOption represents a selectable GPU driver option.
type GPUOption struct {
	Label   string
	Vendor  system.GPUVendor
	Drivers []string
}

// GPUModelImpl holds GPU selection state.
type GPUModelImpl struct {
	detected *system.GPU
	options  []GPUOption
	selected int
}

// NewGPUModel creates a new GPU selection model.
func NewGPUModel() *GPUModelImpl {
	options := []GPUOption{
		{
			Label:   "AMD",
			Vendor:  system.GPUVendorAMD,
			Drivers: []string{"mesa", "vulkan-radeon", "libva-mesa-driver", "mesa-vdpau"},
		},
		{
			Label:   "Intel",
			Vendor:  system.GPUVendorIntel,
			Drivers: []string{"mesa", "vulkan-intel", "intel-media-driver"},
		},
		{
			Label:   "NVIDIA (open)",
			Vendor:  system.GPUVendorNVIDIA,
			Drivers: []string{"nvidia-open", "nvidia-utils", "libva-nvidia-driver"},
		},
		{
			Label:   "Skip GPU drivers",
			Vendor:  system.GPUVendorUnknown,
			Drivers: []string{},
		},
	}

	return &GPUModelImpl{
		options:  options,
		selected: 0,
	}
}

// SetDetectedGPU sets the detected GPU and aligns selection.
func (gm *GPUModelImpl) SetDetectedGPU(gpu *system.GPU) {
	gm.detected = gpu
	if gpu == nil {
		return
	}
	for i, option := range gm.options {
		if option.Vendor == gpu.Vendor() {
			gm.selected = i
			return
		}
	}
}

// DetectedGPU returns the detected GPU.
func (gm *GPUModelImpl) DetectedGPU() *system.GPU {
	return gm.detected
}

// Options returns the selectable options.
func (gm *GPUModelImpl) Options() []GPUOption {
	return gm.options
}

// SelectedIndex returns the current selection index.
func (gm *GPUModelImpl) SelectedIndex() int {
	return gm.selected
}

// SelectedOption returns the currently selected option.
func (gm *GPUModelImpl) SelectedOption() GPUOption {
	if gm.selected < 0 || gm.selected >= len(gm.options) {
		return gm.options[0]
	}
	return gm.options[gm.selected]
}

// MoveUp moves selection up.
func (gm *GPUModelImpl) MoveUp() {
	if gm.selected == 0 {
		gm.selected = len(gm.options) - 1
		return
	}
	gm.selected--
}

// MoveDown moves selection down.
func (gm *GPUModelImpl) MoveDown() {
	gm.selected = (gm.selected + 1) % len(gm.options)
}
