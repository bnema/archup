package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderKernelSelection renders the kernel selection screen.
func RenderKernelSelection(km *models.KernelModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	b.WriteString("\n")
	b.WriteString(title.Render("Kernel Selection"))
	b.WriteString("\n\n")

	if km == nil || len(km.Options()) == 0 {
		b.WriteString(info.Render("No kernel options available."))
		return b.String()
	}

	for i, option := range km.Options() {
		prefix := "  "
		style := lipgloss.NewStyle()
		label := option.Package
		if option.Recommended {
			label = label + " (recommended)"
		}
		if i == km.SelectedIndex() {
			prefix = "> "
			style = active
		}
		b.WriteString(style.Render(prefix + label))
		if option.Description != "" {
			b.WriteString("\n")
			b.WriteString(info.Render("    " + option.Description))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(info.Render("↑/↓ navigate • enter confirm • esc back • ctrl+c quit"))

	return b.String()
}
