package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/bnema/archup/internal/interfaces/tui/models"
)

// HandleFormUpdate processes form input and updates the form model
func HandleFormUpdate(fm *models.FormModelImpl, msg tea.Msg) (*models.FormModelImpl, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			fm.FocusPrevious()
		case "down", "tab":
			fm.FocusNext()
		case "enter":
			if fm.GetFocusIndex() == len(fm.GetFields())-1 {
				// Last field, submit
				fm.SetSubmitted(true)
				fm.ExtractData()
				return fm, nil
			}
			fm.FocusNext()
		case "ctrl+c":
			return fm, tea.Quit
		}
	}

	// Update focused field
	cmd := fm.UpdateInput(msg)
	return fm, cmd
}
