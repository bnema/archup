package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderEncryptionSelection renders the encryption selection screen.
func RenderEncryptionSelection(em *models.EncryptionModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)

	b.WriteString("\n")
	b.WriteString(title.Render("Disk Encryption"))
	b.WriteString("\n\n")

	b.WriteString(info.Render("Choose how to encrypt your installation disk."))
	b.WriteString("\n\n")

	for i, option := range em.Options() {
		prefix := "  "
		style := lipgloss.NewStyle()
		labelStyle := lipgloss.NewStyle()

		if i == em.SelectedIndex() {
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
	b.WriteString(info.Render("↑/↓ to select • enter to confirm • esc to go back"))

	return b.String()
}
