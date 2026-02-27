package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderGPUSelection renders the GPU selection screen.
func RenderGPUSelection(gm *models.GPUModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	b.WriteString("\n")
	b.WriteString(title.Render("GPU Driver Selection"))
	b.WriteString("\n\n")

	if detected := gm.DetectedGPU(); detected != nil {
		b.WriteString(info.Render("Detected: "))
		b.WriteString(detected.Vendor().String())
		b.WriteString(" - ")
		b.WriteString(detected.Model())
		b.WriteString("\n\n")
	}

	for i, option := range gm.Options() {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == gm.SelectedIndex() {
			prefix = "› "
			style = active
		}
		b.WriteString(style.Render(prefix + option.Label))
		if len(option.Drivers) > 0 {
			b.WriteString("\n")
			b.WriteString(info.Render("    " + strings.Join(option.Drivers, ", ")))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(info.Render("↑/↓ navigate • enter confirm • esc back • ctrl+c quit"))

	return b.String()
}
