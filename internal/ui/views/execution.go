package views

import (
	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderExecuting renders the installation progress view
func RenderExecuting(m model.UI) string {
	title := m.Spinner().View() + " Installation in progress..."
	return styles.TitleStyle.Render(title) + "\n\n" + m.Output().View()
}
