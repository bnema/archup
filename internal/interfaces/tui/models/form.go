package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BaseModel provides common functionality for models
type BaseModel struct {
	width  int
	height int
}

// SetSize sets the terminal size for the model
func (bm *BaseModel) SetSize(width, height int) {
	bm.width = width
	bm.height = height
}

// FormData contains the data from the form
type FormData struct {
	Hostname       string
	Username       string
	UserPassword   string
	RootPassword   string
	TargetDisk     string
	EncryptionType string
	Timezone       string
	Locale         string
	Keymap         string
	KernelVariant  string
	AURHelper      string
	Microcode      bool
}

// FormModelImpl implements FormModel interface
type FormModelImpl struct {
	BaseModel

	// Form fields
	hostname     textinput.Model
	username     textinput.Model
	userPassword textinput.Model
	rootPassword textinput.Model
	targetDisk   textinput.Model
	timezone     textinput.Model
	locale       textinput.Model
	keymap       textinput.Model

	// UI state
	focusIndex int
	fields     []textinput.Model
	submitted  bool
	err        error

	// Form data
	data FormData
}

// NewFormModel creates a new form model
func NewFormModel() *FormModelImpl {
	fm := &FormModelImpl{
		fields: []textinput.Model{},
	}

	// Initialize text inputs with labels
	fm.hostname = createTextInput("Hostname", "myarch", "Enter system hostname")
	fm.username = createTextInput("Username", "archuser", "Enter regular user name")
	fm.userPassword = createTextInput("User Password", "", "Enter user password")
	fm.userPassword.EchoMode = textinput.EchoPassword
	fm.rootPassword = createTextInput("Root Password", "", "Enter root password")
	fm.rootPassword.EchoMode = textinput.EchoPassword
	fm.targetDisk = createTextInput("Target Disk", "/dev/sda", "Enter target disk (e.g., /dev/sda)")
	fm.timezone = createTextInput("Timezone", "UTC", "Enter timezone")
	fm.locale = createTextInput("Locale", "en_US.UTF-8", "Enter locale")
	fm.keymap = createTextInput("Keymap", "us", "Enter keymap")

	fm.fields = []textinput.Model{
		fm.hostname, fm.username, fm.userPassword, fm.rootPassword,
		fm.targetDisk, fm.timezone, fm.locale, fm.keymap,
	}

	// Set first field as focused
	fm.fields[0].Focus()

	return fm
}

// Init initializes the form model
func (fm *FormModelImpl) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input and updates the form
// Deprecated: Use handlers.HandleFormUpdate instead
func (fm *FormModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			fm.focusPrevious()
		case "down", "tab":
			fm.focusNext()
		case "enter":
			if fm.focusIndex == len(fm.fields)-1 {
				// Last field, submit
				fm.submitted = true
				fm.extractData()
				return fm, nil
			}
			fm.focusNext()
		case "ctrl+c":
			return fm, tea.Quit
		}
	}

	// Update focused field
	cmd := fm.updateInput(msg)
	return fm, cmd
}

// View renders the form
// Deprecated: Use views.RenderForm instead
func (fm *FormModelImpl) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("ArchUp Installer Configuration"))
	b.WriteString("\n\n")

	labels := []string{
		"Hostname:",
		"Username:",
		"User Password:",
		"Root Password:",
		"Target Disk:",
		"Timezone:",
		"Locale:",
		"Keymap:",
	}

	for i, field := range fm.fields {
		if i == fm.focusIndex {
			// Focused field gets highlighted rendering
			label := lipgloss.NewStyle().
				Width(15).
				Foreground(lipgloss.Color("10")).
				Bold(true).
				Render(labels[i])
			b.WriteString(label)
			b.WriteString(" ")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true).
				Render(field.View()))
		} else {
			// Unfocused field gets dimmed rendering
			label := lipgloss.NewStyle().
				Width(15).
				Foreground(lipgloss.Color("8")).
				Render(labels[i])
			b.WriteString(label)
			b.WriteString(" ")
			b.WriteString(field.View())
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("↑↓ Navigate • Tab/Shift+Tab Switch • Enter Submit • Ctrl+C Quit"))

	if fm.err != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + fm.err.Error()))
	}

	return b.String()
}

// GetData returns the form data
func (fm *FormModelImpl) GetData() FormData {
	return fm.data
}

// SetData sets the form data
func (fm *FormModelImpl) SetData(data FormData) {
	fm.data = data
	fm.hostname.SetValue(data.Hostname)
	fm.username.SetValue(data.Username)
	fm.userPassword.SetValue(data.UserPassword)
	fm.rootPassword.SetValue(data.RootPassword)
	fm.targetDisk.SetValue(data.TargetDisk)
	fm.timezone.SetValue(data.Timezone)
	fm.locale.SetValue(data.Locale)
	fm.keymap.SetValue(data.Keymap)
}

// GetFields returns the form fields
func (fm *FormModelImpl) GetFields() []textinput.Model {
	return fm.fields
}

// GetFocusIndex returns the currently focused field index
func (fm *FormModelImpl) GetFocusIndex() int {
	return fm.focusIndex
}

// GetError returns the form error if any
func (fm *FormModelImpl) GetError() error {
	return fm.err
}

// SetError sets an error on the form
func (fm *FormModelImpl) SetError(err error) {
	fm.err = err
}

// SetSubmitted marks the form as submitted
func (fm *FormModelImpl) SetSubmitted(submitted bool) {
	fm.submitted = submitted
}

// IsSubmitted returns whether the form was submitted
func (fm *FormModelImpl) IsSubmitted() bool {
	return fm.submitted
}

// FocusNext moves focus to the next field
func (fm *FormModelImpl) FocusNext() {
	fm.focusNext()
}

// FocusPrevious moves focus to the previous field
func (fm *FormModelImpl) FocusPrevious() {
	fm.focusPrevious()
}

// UpdateInput updates the currently focused input field
func (fm *FormModelImpl) UpdateInput(msg tea.Msg) tea.Cmd {
	return fm.updateInput(msg)
}

// ExtractData extracts form values into the data struct
func (fm *FormModelImpl) ExtractData() {
	fm.extractData()
}

// Helper methods

func (fm *FormModelImpl) focusNext() {
	fm.fields[fm.focusIndex].Blur()
	fm.focusIndex = (fm.focusIndex + 1) % len(fm.fields)
	fm.fields[fm.focusIndex].Focus()
}

func (fm *FormModelImpl) focusPrevious() {
	fm.fields[fm.focusIndex].Blur()
	fm.focusIndex--
	if fm.focusIndex < 0 {
		fm.focusIndex = len(fm.fields) - 1
	}
	fm.fields[fm.focusIndex].Focus()
}

func (fm *FormModelImpl) updateInput(msg tea.Msg) tea.Cmd {
	_, cmd := fm.fields[fm.focusIndex].Update(msg)
	return cmd
}

func (fm *FormModelImpl) extractData() {
	fm.data = FormData{
		Hostname:     fm.hostname.Value(),
		Username:     fm.username.Value(),
		UserPassword: fm.userPassword.Value(),
		RootPassword: fm.rootPassword.Value(),
		TargetDisk:   fm.targetDisk.Value(),
		Timezone:     fm.timezone.Value(),
		Locale:       fm.locale.Value(),
		Keymap:       fm.keymap.Value(),
	}
}

// createTextInput creates a configured text input field
func createTextInput(label, placeholder, help string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 30
	return ti
}
