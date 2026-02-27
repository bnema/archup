package models

// DankLinuxOption represents a yes/no option for installing Dank Linux
type DankLinuxOption struct {
	Value       bool
	Label       string
	Description string
}

// DankLinuxModelImpl is the model for the Dank Linux opt-in toggle
type DankLinuxModelImpl struct {
	options  []DankLinuxOption
	selected int
}

// NewDankLinuxModel creates a new Dank Linux opt-in selection model.
func NewDankLinuxModel() *DankLinuxModelImpl {
	return &DankLinuxModelImpl{
		options: []DankLinuxOption{
			{
				Value:       false,
				Label:       "No, keep barebone Bash",
				Description: "Barebone system with modern CLI tools. Install Dank Linux later with: curl -fsSL https://install.danklinux.com | sh",
			},
			{
				Value:       true,
				Label:       "Yes, install Dank Linux",
				Description: "Full Wayland desktop: niri or Hyprland + DankMaterialShell + Ghostty + auto-theming. Runs on first boot.",
			},
		},
		selected: 0,
	}
}

// Options returns the selectable options.
func (d *DankLinuxModelImpl) Options() []DankLinuxOption {
	return d.options
}

// SelectedIndex returns the current selection index.
func (d *DankLinuxModelImpl) SelectedIndex() int {
	return d.selected
}

// SelectedOption returns the currently selected option.
func (d *DankLinuxModelImpl) SelectedOption() DankLinuxOption {
	if len(d.options) == 0 {
		return DankLinuxOption{}
	}
	if d.selected < 0 || d.selected >= len(d.options) {
		return d.options[0]
	}
	return d.options[d.selected]
}

// MoveUp moves selection up, wrapping to the last option.
func (d *DankLinuxModelImpl) MoveUp() {
	if d.selected > 0 {
		d.selected--
	} else {
		d.selected = len(d.options) - 1
	}
}

// MoveDown moves selection down, wrapping to the first option.
func (d *DankLinuxModelImpl) MoveDown() {
	if d.selected < len(d.options)-1 {
		d.selected++
	} else {
		d.selected = 0
	}
}
