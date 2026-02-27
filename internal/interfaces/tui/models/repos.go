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
// Navigation: ↑/↓ moves a single cursor across all items (crossing sections).
// ←/→ selects the item under the cursor (radio within its section).
type ReposModelImpl struct {
	aurOptions      []AURHelperOption
	aurSelected     int // which AUR item is selected (confirmed)
	chaoticOptions  []ChaoticOption
	chaoticSelected int // which Chaotic item is selected (confirmed)
	cursor          int // flat index across all items (0..totalItems-1)
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
		cursor:          0,
	}
}

// totalItems returns the total number of items across all sections.
func (rm *ReposModelImpl) totalItems() int {
	return len(rm.aurOptions) + len(rm.chaoticOptions)
}

// CursorIndex returns the current flat cursor position.
func (rm *ReposModelImpl) CursorIndex() int { return rm.cursor }

// AUROptions returns the selectable AUR helper options.
func (rm *ReposModelImpl) AUROptions() []AURHelperOption { return rm.aurOptions }

// AURSelectedIndex returns the confirmed AUR selection index.
func (rm *ReposModelImpl) AURSelectedIndex() int { return rm.aurSelected }

// ChaoticOptions returns the selectable Chaotic-AUR options.
func (rm *ReposModelImpl) ChaoticOptions() []ChaoticOption { return rm.chaoticOptions }

// ChaoticSelectedIndex returns the confirmed Chaotic-AUR selection index.
func (rm *ReposModelImpl) ChaoticSelectedIndex() int { return rm.chaoticSelected }

// FocusSection returns which section the cursor is in (0=AUR, 1=Chaotic).
// Kept for view compatibility.
func (rm *ReposModelImpl) FocusSection() int {
	if rm.cursor < len(rm.aurOptions) {
		return 0
	}
	return 1
}

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

// MoveUp moves the cursor up (crosses section boundaries, no wrap).
func (rm *ReposModelImpl) MoveUp() {
	if rm.cursor > 0 {
		rm.cursor--
	}
}

// MoveDown moves the cursor down (crosses section boundaries, no wrap).
func (rm *ReposModelImpl) MoveDown() {
	if rm.cursor < rm.totalItems()-1 {
		rm.cursor++
	}
}

// Select selects (radio) the item currently under the cursor within its section.
func (rm *ReposModelImpl) Select() {
	if rm.cursor < len(rm.aurOptions) {
		rm.aurSelected = rm.cursor
	} else {
		rm.chaoticSelected = rm.cursor - len(rm.aurOptions)
	}
}

// NextSection moves focus to the next section (kept for compatibility).
func (rm *ReposModelImpl) NextSection() {
	rm.focusToSection(rm.FocusSection() + 1)
}

// PrevSection moves focus to the previous section (kept for compatibility).
func (rm *ReposModelImpl) PrevSection() {
	rm.focusToSection(rm.FocusSection() - 1)
}

func (rm *ReposModelImpl) focusToSection(s int) {
	if s <= 0 {
		rm.cursor = 0
	} else {
		rm.cursor = len(rm.aurOptions)
	}
}

// SectionCount returns the total number of sections.
func (rm *ReposModelImpl) SectionCount() int { return 2 }
