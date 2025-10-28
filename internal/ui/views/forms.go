package views

import (
	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderForm renders any form view (container centered, content left-aligned)
func RenderForm(m model.UI) string {
	if m.CurrentForm() == nil {
		return "Loading form..."
	}

	header := CenterText(m.RenderPhaseHeader(), m.Width())
	form := m.CurrentForm().View()
	content := styles.FormContainerStyle.Render(header + "\n\n" + form)

	// Center the container on screen, but keep form content left-aligned
	return lipgloss.Place(
		m.Width(),
		m.Height(),
		lipgloss.Center,
		lipgloss.Top,
		content,
		lipgloss.WithWhitespaceChars(" "),
	)
}
