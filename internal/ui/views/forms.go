package views

import (
	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderForm renders any form view (left-aligned with padding)
func RenderForm(m model.UI) string {
	if m.CurrentForm() == nil {
		return "Loading form..."
	}

	header := CenterText(m.RenderPhaseHeader(), m.Width())
	form := m.CurrentForm().View()

	return styles.FormContainerStyle.Render(header + "\n\n" + form)
}
