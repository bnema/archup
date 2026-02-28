package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderReposSelection renders the AUR helper selection screen.
// Chaotic-AUR is always enabled and not shown as an option.
func RenderReposSelection(rm *models.ReposModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	section := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	selectedMark := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	normalMark := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString("\n")
	b.WriteString(title.Render("Repository Configuration"))
	b.WriteString("\n\n")

	b.WriteString(section.Render("AUR Helper"))
	b.WriteString("\n")

	for i, option := range rm.AUROptions() {
		isCursor := rm.CursorIndex() == i
		isSelected := rm.AURSelectedIndex() == i

		if isSelected {
			b.WriteString(selectedMark.Render(" [x] "))
		} else {
			b.WriteString(normalMark.Render(" [ ] "))
		}

		label := option.Label
		if isCursor {
			b.WriteString(cursorStyle.Render("> " + label))
		} else {
			b.WriteString(label)
		}
		b.WriteString("\n")
		b.WriteString(desc.Render("      " + option.Description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(desc.Render("  Chaotic-AUR is always enabled (pre-built AUR packages)"))
	b.WriteString("\n\n")
	b.WriteString(info.Render("↑/↓ move • ←/→ or space select • enter confirm • esc back"))

	return b.String()
}
