package components

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/bnema/archup/internal/ui/styles"
)

// FormBuilder provides reusable form components
type FormBuilder struct {
	accessible bool
	width      int
}

// NewFormBuilder creates a new form builder with specified width
func NewFormBuilder(accessible bool, width int) *FormBuilder {
	return &FormBuilder{
		accessible: accessible,
		width:      width,
	}
}

// TextInput creates a text input field
func (fb *FormBuilder) TextInput(title, prompt string, value *string, validator func(string) error) *huh.Input {
	input := huh.NewInput().
		Title(title).
		Prompt(prompt).
		Value(value)

	if validator != nil {
		input.Validate(validator)
	}

	return input
}

// Select creates a single-selection field
func (fb *FormBuilder) Select(title string, options []string, value *string) *huh.Select[string] {
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
	}

	return huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(value)
}

// MultiSelect creates a multi-selection field
func (fb *FormBuilder) MultiSelect(title string, options []string, value *[]string, limit int) *huh.MultiSelect[string] {
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
	}

	ms := huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Value(value)

	if limit > 0 {
		ms.Limit(limit)
	}

	return ms
}

// Confirm creates a yes/no confirmation with left-aligned buttons
func (fb *FormBuilder) Confirm(title, affirmative, negative string, value *bool) *huh.Confirm {
	return huh.NewConfirm().
		Title(title).
		Affirmative(affirmative).
		Negative(negative).
		Value(value).
		WithButtonAlignment(lipgloss.Left)
}

// CreateForm creates a form with accessibility settings and custom theme
func (fb *FormBuilder) CreateForm(groups ...*huh.Group) *huh.Form {
	formWidth := fb.width
	if formWidth > 76 {
		formWidth = 76
	}
	if formWidth < 40 {
		formWidth = 40
	}

	form := huh.NewForm(groups...).
		WithTheme(styles.HuhTheme()).
		WithKeyMap(CreateTTYKeyMap()).
		WithWidth(formWidth)

	if fb.accessible {
		form.WithAccessible(true)
	}
	return form
}

// PasswordInput creates a password input field
func (fb *FormBuilder) PasswordInput(title, prompt string, value *string, validator func(string) error) *huh.Input {
	input := fb.TextInput(title, prompt, value, validator)
	input.EchoMode(huh.EchoModePassword)
	return input
}

// SelectWithOptions creates a select field with option objects (for labels different from values)
func (fb *FormBuilder) SelectWithOptions(title string, options []huh.Option[string], value *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(value)
}

// CreateOption creates a labeled option for Select fields (DRY helper)
func CreateOption(label, description, value string) huh.Option[string] {
	switch {
	case description != "":
		return huh.NewOption(label+" - "+description, value)
	default:
		return huh.NewOption(label, value)
	}
}

// CreateKernelOptions creates kernel selection options with descriptions
func CreateKernelOptions() []huh.Option[string] {
	return []huh.Option[string]{
		CreateOption("linux", "Stable mainline kernel (recommended)", "linux"),
		CreateOption("linux-lts", "Long-term support (maximum stability)", "linux-lts"),
		CreateOption("linux-zen", "Performance-optimized for general use", "linux-zen"),
		CreateOption("linux-hardened", "Security-focused kernel", "linux-hardened"),
		CreateOption("linux-cachyos", "Gaming-optimized", "linux-cachyos"),
	}
}

// CreateAMDPStateOptions creates AMD P-State mode options with descriptions
func CreateAMDPStateOptions(modes []string) []huh.Option[string] {
	var options []huh.Option[string]

	for _, mode := range modes {
		var description string
		switch mode {
		case "active":
			description = "Best performance (recommended for desktop/gaming)"
		case "guided":
			description = "Balanced performance and efficiency (recommended for laptops)"
		case "passive":
			description = "Maximum compatibility (older CPUs)"
		}

		switch {
		case description != "":
			options = append(options, huh.NewOption(mode+" - "+description, mode))
		default:
			options = append(options, huh.NewOption(mode, mode))
		}
	}

	return options
}
