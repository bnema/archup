package models

// AURHelperOption represents a selectable AUR helper option.
type AURHelperOption struct {
	Value       string
	Label       string
	Description string
	Recommended bool
}

// ReposModelImpl holds AUR helper selection state.
// Chaotic-AUR is always enabled — not user-configurable.
type ReposModelImpl struct {
	aurOptions  []AURHelperOption
	aurSelected int // which AUR item is selected (confirmed)
	cursor      int // flat index across all items
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

	return &ReposModelImpl{
		aurOptions:  aurOptions,
		aurSelected: aurSelected,
		cursor:      0,
	}
}

// CursorIndex returns the current flat cursor position.
func (rm *ReposModelImpl) CursorIndex() int { return rm.cursor }

// AUROptions returns the selectable AUR helper options.
func (rm *ReposModelImpl) AUROptions() []AURHelperOption { return rm.aurOptions }

// AURSelectedIndex returns the confirmed AUR selection index.
func (rm *ReposModelImpl) AURSelectedIndex() int { return rm.aurSelected }

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

// MoveUp moves the cursor up.
func (rm *ReposModelImpl) MoveUp() {
	if rm.cursor > 0 {
		rm.cursor--
	}
}

// MoveDown moves the cursor down.
func (rm *ReposModelImpl) MoveDown() {
	if rm.cursor < len(rm.aurOptions)-1 {
		rm.cursor++
	}
}

// Select confirms the item currently under the cursor.
func (rm *ReposModelImpl) Select() {
	if rm.cursor < len(rm.aurOptions) {
		rm.aurSelected = rm.cursor
	}
}
