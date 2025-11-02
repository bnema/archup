package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

// CreateTTYKeyMap creates a custom keymap with TTY-friendly alternatives to Shift+Tab
// This is needed for QEMU console and other limited terminal environments where
// Shift+Tab is not properly transmitted.
func CreateTTYKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()

	// Add Ctrl+P as alternative to Shift+Tab for "previous" navigation
	// These work better in QEMU console and limited TTY environments
	km.Input.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.Text.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.Select.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.MultiSelect.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.Note.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.Confirm.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)
	km.FilePicker.Prev = key.NewBinding(
		key.WithKeys("shift+tab", "ctrl+p"),
		key.WithHelp("ctrl+p/shift+tab", "back"),
	)

	return km
}
