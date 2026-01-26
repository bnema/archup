package models

import "github.com/bnema/archup/internal/domain/packages"

// KernelOption represents a selectable kernel option.
type KernelOption struct {
	Variant     packages.KernelVariant
	Package     string
	Description string
	Recommended bool
}

// KernelModelImpl holds kernel selection state.
type KernelModelImpl struct {
	options  []KernelOption
	selected int
}

// NewKernelModel creates a new kernel selection model.
func NewKernelModel() *KernelModelImpl {
	options := []KernelOption{
		{
			Variant:     packages.KernelStable,
			Package:     packages.KernelStable.String(),
			Description: "Stable mainline kernel",
			Recommended: true,
		},
		{
			Variant:     packages.KernelLTS,
			Package:     packages.KernelLTS.String(),
			Description: "Long-term support (maximum stability)",
		},
		{
			Variant:     packages.KernelZen,
			Package:     packages.KernelZen.String(),
			Description: "Performance-optimized for general use",
		},
		{
			Variant:     packages.KernelHardened,
			Package:     packages.KernelHardened.String(),
			Description: "Security-focused kernel",
		},
		{
			Variant:     packages.KernelCachyOS,
			Package:     packages.KernelCachyOS.String(),
			Description: "Gaming-optimized",
		},
	}

	selected := 0
	for i, option := range options {
		if option.Recommended {
			selected = i
			break
		}
	}

	return &KernelModelImpl{
		options:  options,
		selected: selected,
	}
}

// Options returns the selectable kernel options.
func (km *KernelModelImpl) Options() []KernelOption {
	return km.options
}

// SelectedIndex returns the current selection index.
func (km *KernelModelImpl) SelectedIndex() int {
	return km.selected
}

// SelectedOption returns the currently selected kernel option.
func (km *KernelModelImpl) SelectedOption() KernelOption {
	if len(km.options) == 0 {
		return KernelOption{}
	}
	if km.selected < 0 || km.selected >= len(km.options) {
		return km.options[0]
	}
	return km.options[km.selected]
}

// MoveUp moves selection up.
func (km *KernelModelImpl) MoveUp() {
	if len(km.options) == 0 {
		return
	}
	if km.selected == 0 {
		km.selected = len(km.options) - 1
		return
	}
	km.selected--
}

// MoveDown moves selection down.
func (km *KernelModelImpl) MoveDown() {
	if len(km.options) == 0 {
		return
	}
	km.selected = (km.selected + 1) % len(km.options)
}

// SetSelectedPackage selects a kernel by package name.
func (km *KernelModelImpl) SetSelectedPackage(pkg string) {
	if pkg == "" {
		return
	}
	for i, option := range km.options {
		if option.Package == pkg {
			km.selected = i
			return
		}
	}
}
