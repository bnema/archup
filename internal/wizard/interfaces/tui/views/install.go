package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderInstall renders a placeholder install screen.
func RenderInstall() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString("\n")
	b.WriteString(title.Render("Installing"))
	b.WriteString("\n\n")
	b.WriteString("Wizard install is not implemented yet.\n\n")
	b.WriteString(muted.Render("Running setup..."))

	return b.String()
}
