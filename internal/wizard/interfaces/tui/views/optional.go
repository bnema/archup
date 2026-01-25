package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderOptional renders optional component toggles.
func RenderOptional(config domain.DesktopConfig) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	cliphistStatus := "off"
	if config.InstallCliphist {
		cliphistStatus = "on"
	}

	b.WriteString("\n")
	b.WriteString(title.Render("Optional Components"))
	b.WriteString("\n\n")
	b.WriteString("Cliphist: " + cliphistStatus + "\n\n")
	b.WriteString(muted.Render("Space Toggle • Enter Next"))

	return b.String()
}
