package models

import legacysystem "github.com/bnema/archup/internal/system"

// DiskOption represents a selectable disk option.
type DiskOption struct {
	Label string
	Path  string
	Size  string
	Model string
}

// DiskModelImpl holds disk selection state.
type DiskModelImpl struct {
	disks    []legacysystem.Disk
	options  []DiskOption
	selected int
	err      error
}

// NewDiskModel creates a new disk selection model.
func NewDiskModel() *DiskModelImpl {
	return &DiskModelImpl{
		options:  []DiskOption{},
		selected: 0,
	}
}

// SetDisks populates the disk options from detected disks.
func (dm *DiskModelImpl) SetDisks(disks []legacysystem.Disk) {
	dm.disks = disks
	dm.options = make([]DiskOption, 0, len(disks))

	for _, disk := range disks {
		label := disk.Path + " (" + disk.Size + ")"
		if disk.Model != "" {
			label += " " + disk.Model
		}
		if disk.Serial != "" {
			label += " [" + disk.Serial + "]"
		}
		if disk.Vendor != "" && disk.Model == "" {
			label += " " + disk.Vendor
		}

		dm.options = append(dm.options, DiskOption{
			Label: label,
			Path:  disk.Path,
			Size:  disk.Size,
			Model: disk.Model,
		})
	}
}

// SetError sets an error that occurred during disk detection.
func (dm *DiskModelImpl) SetError(err error) {
	dm.err = err
}

// Error returns any error from disk detection.
func (dm *DiskModelImpl) Error() error {
	return dm.err
}

// Options returns the selectable disk options.
func (dm *DiskModelImpl) Options() []DiskOption {
	return dm.options
}

// SelectedIndex returns the current selection index.
func (dm *DiskModelImpl) SelectedIndex() int {
	return dm.selected
}

// SelectedOption returns the currently selected disk option.
func (dm *DiskModelImpl) SelectedOption() DiskOption {
	if len(dm.options) == 0 {
		return DiskOption{Label: "No disks found", Path: ""}
	}
	if dm.selected < 0 || dm.selected >= len(dm.options) {
		return dm.options[0]
	}
	return dm.options[dm.selected]
}

// MoveUp moves selection up.
func (dm *DiskModelImpl) MoveUp() {
	if len(dm.options) == 0 {
		return
	}
	if dm.selected == 0 {
		dm.selected = len(dm.options) - 1
		return
	}
	dm.selected--
}

// MoveDown moves selection down.
func (dm *DiskModelImpl) MoveDown() {
	if len(dm.options) == 0 {
		return
	}
	dm.selected = (dm.selected + 1) % len(dm.options)
}

// HasDisks returns true if any disks were detected.
func (dm *DiskModelImpl) HasDisks() bool {
	return len(dm.options) > 0
}
