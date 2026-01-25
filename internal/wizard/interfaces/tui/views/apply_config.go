package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderApplyConfig renders the config apply step.
func RenderApplyConfig() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString("\n")
	b.WriteString(title.Render("Applying Configuration"))
	b.WriteString("\n\n")
	b.WriteString("Writing compositor config and enabling services...\n\n")
	b.WriteString(muted.Render("Please wait."))

	return b.String()
}
