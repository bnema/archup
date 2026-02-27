package models

// AURHelperOption represents a selectable AUR helper option.
type AURHelperOption struct {
	Value       string
	Label       string
	Description string
	Recommended bool
}

// ChaoticOption represents a selectable Chaotic-AUR option.
type ChaoticOption struct {
	Enabled     bool
	Label       string
	Description string
	Recommended bool
}

// ReposModelImpl holds AUR helper and Chaotic-AUR selection state.
type ReposModelImpl struct {
	aurOptions      []AURHelperOption
	aurSelected     int
	chaoticOptions  []ChaoticOption
	chaoticSelected int
	// focusSection: 0 = AUR helper, 1 = Chaotic-AUR
	focusSection int
}

// NewReposModel creates a new repos selection model.
func NewReposModel() *ReposModelImpl {
	aurOptions := []AURHelperOption{
		{Value: "paru", Label: "paru", Description: "Fast AUR helper written in Rust (recommended)", Recommended: true},
		{Value: "yay", Label: "yay", Description: "Yet Another Yogurt - AUR helper written in Go"},
	}
	aurSelected := 0
	for i, o := range aurOptions {
		if o.Recommended {
			aurSelected = i
			break
		}
	}

	chaoticOptions := []ChaoticOption{
		{Enabled: true, Label: "Enabled", Description: "Pre-built AUR packages; faster installs (recommended)", Recommended: true},
		{Enabled: false, Label: "Disabled", Description: "Build AUR packages from source"},
	}
	chaoticSelected := 0
	for i, o := range chaoticOptions {
		if o.Recommended {
			chaoticSelected = i
			break
		}
	}

	return &ReposModelImpl{
		aurOptions:      aurOptions,
		aurSelected:     aurSelected,
		chaoticOptions:  chaoticOptions,
		chaoticSelected: chaoticSelected,
		focusSection:    0,
	}
}

// AUROptions returns the selectable AUR helper options.
func (rm *ReposModelImpl) AUROptions() []AURHelperOption { return rm.aurOptions }

// AURSelectedIndex returns the current AUR selection index.
func (rm *ReposModelImpl) AURSelectedIndex() int { return rm.aurSelected }

// ChaoticOptions returns the selectable Chaotic-AUR options.
func (rm *ReposModelImpl) ChaoticOptions() []ChaoticOption { return rm.chaoticOptions }

// ChaoticSelectedIndex returns the current Chaotic-AUR selection index.
func (rm *ReposModelImpl) ChaoticSelectedIndex() int { return rm.chaoticSelected }

// FocusSection returns the currently focused section (0=AUR, 1=Chaotic).
func (rm *ReposModelImpl) FocusSection() int { return rm.focusSection }

// SelectedAURHelper returns the selected AUR helper value string.
func (rm *ReposModelImpl) SelectedAURHelper() string {
	if len(rm.aurOptions) == 0 {
		return "paru"
	}
	if rm.aurSelected < 0 || rm.aurSelected >= len(rm.aurOptions) {
		return rm.aurOptions[0].Value
	}
	return rm.aurOptions[rm.aurSelected].Value
}

// SelectedChaoticEnabled returns whether Chaotic-AUR is enabled.
func (rm *ReposModelImpl) SelectedChaoticEnabled() bool {
	if len(rm.chaoticOptions) == 0 {
		return true
	}
	if rm.chaoticSelected < 0 || rm.chaoticSelected >= len(rm.chaoticOptions) {
		return rm.chaoticOptions[0].Enabled
	}
	return rm.chaoticOptions[rm.chaoticSelected].Enabled
}

// MoveUp moves selection up in the current section (wraps).
func (rm *ReposModelImpl) MoveUp() {
	if rm.focusSection == 0 {
		if len(rm.aurOptions) == 0 {
			return
		}
		if rm.aurSelected == 0 {
			rm.aurSelected = len(rm.aurOptions) - 1
		} else {
			rm.aurSelected--
		}
	} else {
		if len(rm.chaoticOptions) == 0 {
			return
		}
		if rm.chaoticSelected == 0 {
			rm.chaoticSelected = len(rm.chaoticOptions) - 1
		} else {
			rm.chaoticSelected--
		}
	}
}

// MoveDown moves selection down in the current section (wraps).
func (rm *ReposModelImpl) MoveDown() {
	if rm.focusSection == 0 {
		if len(rm.aurOptions) == 0 {
			return
		}
		rm.aurSelected = (rm.aurSelected + 1) % len(rm.aurOptions)
	} else {
		if len(rm.chaoticOptions) == 0 {
			return
		}
		rm.chaoticSelected = (rm.chaoticSelected + 1) % len(rm.chaoticOptions)
	}
}

// NextSection moves focus to the next section.
func (rm *ReposModelImpl) NextSection() {
	rm.focusSection = (rm.focusSection + 1) % 2
}

// PrevSection moves focus to the previous section.
func (rm *ReposModelImpl) PrevSection() {
	rm.focusSection = (rm.focusSection - 1 + 2) % 2
}
