package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderWelcome renders the wizard welcome screen.
func RenderWelcome(config domain.DesktopConfig) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString("\n")
	b.WriteString(title.Render("ArchUp Desktop Wizard"))
	b.WriteString("\n\n")
	b.WriteString("This wizard will install a Wayland desktop and apply defaults.\n")
	b.WriteString("You can re-run it safely to reconfigure.\n\n")
	b.WriteString(muted.Render("Default compositor: " + string(config.Compositor) + "\n"))
	b.WriteString(muted.Render("Press Enter to begin or Ctrl+C to exit."))
	return b.String()
}
