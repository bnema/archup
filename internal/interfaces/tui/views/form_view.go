package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderForm renders the form model to a styled string
func RenderForm(fm *models.FormModelImpl) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("ArchUp Installer Configuration"))
	b.WriteString("\n\n")

	labels := []string{
		"Hostname:",
		"Username:",
		"Email(opt):",
		"Password:",
		"Timezone:",
		"Locale:",
		"Keymap:",
	}

	fields := fm.GetFields()
	focusIndex := fm.GetFocusIndex()

	for i, field := range fields {
		if i == focusIndex {
			// Focused field gets highlighted rendering
			label := lipgloss.NewStyle().
				Width(15).
				Foreground(lipgloss.Color("10")).
				Bold(true).
				Render(labels[i])
			b.WriteString(label)
			b.WriteString(" ")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true).
				Render(field.View()))
		} else {
			// Unfocused field gets dimmed rendering
			label := lipgloss.NewStyle().
				Width(15).
				Foreground(lipgloss.Color("8")).
				Render(labels[i])
			b.WriteString(label)
			b.WriteString(" ")
			b.WriteString(field.View())
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("↑↓ Navigate • Tab/Shift+Tab Switch • Enter Submit • Ctrl+C Quit"))

	if err := fm.GetError(); err != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error: " + err.Error()))
	}

	return b.String()
}
