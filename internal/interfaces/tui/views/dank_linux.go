package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderDankLinuxSelection renders the Dank Linux opt-in screen.
func RenderDankLinuxSelection(dm *models.DankLinuxModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)

	b.WriteString("\n")
	b.WriteString(title.Render("Dank Linux Desktop"))
	b.WriteString("\n\n")

	b.WriteString(info.Render("Optional: install Dank Linux — a full Wayland desktop suite."))
	b.WriteString("\n")
	b.WriteString(info.Render("Includes: niri or Hyprland, DankMaterialShell, Ghostty, and auto-theming."))
	b.WriteString("\n\n")

	for i, option := range dm.Options() {
		prefix := "  "
		style := lipgloss.NewStyle()
		labelStyle := lipgloss.NewStyle()

		if i == dm.SelectedIndex() {
			prefix = "> "
			style = active
			labelStyle = active
		}

		b.WriteString(style.Render(prefix))
		b.WriteString(labelStyle.Render(option.Label))
		b.WriteString("\n")
		b.WriteString(desc.Render("    " + option.Description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(info.Render("↑/↓ to select • enter to confirm"))

	return b.String()
}
