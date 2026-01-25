package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderComplete renders completion screen.
func RenderComplete(config domain.DesktopConfig, errMsg string) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	alert := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	b.WriteString("\n")
	if errMsg != "" {
		b.WriteString(alert.Render("Wizard finished with errors"))
		b.WriteString("\n\n")
		b.WriteString(alert.Render(errMsg))
		b.WriteString("\n\n")
	} else {
		b.WriteString(title.Render("Wizard Complete"))
		b.WriteString("\n\n")
		b.WriteString("Desktop setup started for: " + string(config.Compositor))
		b.WriteString("\n\n")
	}

	b.WriteString(muted.Render("Press Enter to exit."))

	return b.String()
}
