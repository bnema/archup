package models

import legacysystem "github.com/bnema/archup/internal/system"

// AMDPStateOption represents a selectable AMD P-State mode.
type AMDPStateOption struct {
	Mode        legacysystem.AMDPStateMode
	Label       string
	Recommended bool
}

// AMDPStateModelImpl holds AMD P-State selection state.
type AMDPStateModelImpl struct {
	cpuInfo  *legacysystem.CPUInfo
	options  []AMDPStateOption
	selected int
}

// NewAMDPStateModel creates a new AMD P-State selection model.
func NewAMDPStateModel() *AMDPStateModelImpl {
	return &AMDPStateModelImpl{
		options:  []AMDPStateOption{},
		selected: 0,
	}
}

// SetCPUInfo sets CPU info and rebuilds options.
func (am *AMDPStateModelImpl) SetCPUInfo(info *legacysystem.CPUInfo) {
	am.cpuInfo = info
	am.options = []AMDPStateOption{}
	am.selected = 0

	if info == nil {
		return
	}

	for _, mode := range info.AMDPStateModes {
		desc := legacysystem.GetPStateModeDescription(mode)
		label := string(mode)
		if desc != "" {
			label = label + " - " + desc
		}
		option := AMDPStateOption{
			Mode:        mode,
			Label:       label,
			Recommended: mode == info.RecommendedPStateMode,
		}
		am.options = append(am.options, option)
	}

	for i, option := range am.options {
		if option.Recommended {
			am.selected = i
			break
		}
	}
}

// CPUInfo returns detected CPU info.
func (am *AMDPStateModelImpl) CPUInfo() *legacysystem.CPUInfo {
	return am.cpuInfo
}

// Options returns the selectable options.
func (am *AMDPStateModelImpl) Options() []AMDPStateOption {
	return am.options
}

// SelectedIndex returns the current selection index.
func (am *AMDPStateModelImpl) SelectedIndex() int {
	return am.selected
}

// SelectedOption returns the currently selected option.
func (am *AMDPStateModelImpl) SelectedOption() AMDPStateOption {
	if len(am.options) == 0 {
		return AMDPStateOption{}
	}
	if am.selected < 0 || am.selected >= len(am.options) {
		return am.options[0]
	}
	return am.options[am.selected]
}

// MoveUp moves selection up.
func (am *AMDPStateModelImpl) MoveUp() {
	if len(am.options) == 0 {
		return
	}
	if am.selected == 0 {
		am.selected = len(am.options) - 1
		return
	}
	am.selected--
}

// MoveDown moves selection down.
func (am *AMDPStateModelImpl) MoveDown() {
	if len(am.options) == 0 {
		return
	}
	am.selected = (am.selected + 1) % len(am.options)
}

// ShouldPrompt returns true if AMD P-State selection should be shown.
func (am *AMDPStateModelImpl) ShouldPrompt() bool {
	if am.cpuInfo == nil {
		return false
	}
	if am.cpuInfo.Vendor != legacysystem.CPUVendorAMD {
		return false
	}
	return len(am.options) > 0
}
