package models

// EncryptionOption represents a selectable encryption option.
type EncryptionOption struct {
	Value       string
	Label       string
	Description string
	Recommended bool
}

// EncryptionModelImpl holds encryption selection state.
type EncryptionModelImpl struct {
	options  []EncryptionOption
	selected int
}

// NewEncryptionModel creates a new encryption selection model.
func NewEncryptionModel() *EncryptionModelImpl {
	options := []EncryptionOption{
		{Value: "luks", Label: "LUKS2", Description: "Full disk encryption (recommended)", Recommended: true},
		{Value: "none", Label: "None", Description: "No encryption"},
	}
	selected := 0
	for i, o := range options {
		if o.Recommended {
			selected = i
			break
		}
	}
	return &EncryptionModelImpl{options: options, selected: selected}
}

// Options returns the selectable encryption options.
func (em *EncryptionModelImpl) Options() []EncryptionOption { return em.options }

// SelectedIndex returns the current selection index.
func (em *EncryptionModelImpl) SelectedIndex() int { return em.selected }

// SelectedOption returns the currently selected option.
func (em *EncryptionModelImpl) SelectedOption() EncryptionOption {
	if len(em.options) == 0 {
		return EncryptionOption{}
	}
	if em.selected < 0 || em.selected >= len(em.options) {
		return em.options[0]
	}
	return em.options[em.selected]
}

// MoveUp moves selection up (wraps).
func (em *EncryptionModelImpl) MoveUp() {
	if len(em.options) == 0 {
		return
	}
	if em.selected == 0 {
		em.selected = len(em.options) - 1
		return
	}
	em.selected--
}

// MoveDown moves selection down (wraps).
func (em *EncryptionModelImpl) MoveDown() {
	if len(em.options) == 0 {
		return
	}
	em.selected = (em.selected + 1) % len(em.options)
}

// SetSelectedValue selects an option by value string.
func (em *EncryptionModelImpl) SetSelectedValue(v string) {
	for i, o := range em.options {
		if o.Value == v {
			em.selected = i
			return
		}
	}
}
