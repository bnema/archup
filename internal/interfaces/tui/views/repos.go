package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderReposSelection renders the AUR helper and Chaotic-AUR selection screen.
func RenderReposSelection(rm *models.ReposModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	section := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	activeSection := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)

	b.WriteString("\n")
	b.WriteString(title.Render("Repository Configuration"))
	b.WriteString("\n\n")

	// AUR Helper section
	aurHeaderStyle := section
	if rm.FocusSection() == 0 {
		aurHeaderStyle = activeSection
	}
	b.WriteString(aurHeaderStyle.Render("AUR Helper"))
	b.WriteString("\n")

	for i, option := range rm.AUROptions() {
		prefix := "  "
		style := lipgloss.NewStyle()
		labelStyle := lipgloss.NewStyle()

		if rm.FocusSection() == 0 && i == rm.AURSelectedIndex() {
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

	// Chaotic-AUR section
	chaoticHeaderStyle := section
	if rm.FocusSection() == 1 {
		chaoticHeaderStyle = activeSection
	}
	b.WriteString(chaoticHeaderStyle.Render("Chaotic-AUR"))
	b.WriteString("\n")

	for i, option := range rm.ChaoticOptions() {
		prefix := "  "
		style := lipgloss.NewStyle()
		labelStyle := lipgloss.NewStyle()

		if rm.FocusSection() == 1 && i == rm.ChaoticSelectedIndex() {
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
	b.WriteString(info.Render("↑/↓ to select • tab to switch section • enter to confirm • esc to go back"))

	return b.String()
}
