package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderDiskSelection renders the disk selection screen.
func RenderDiskSelection(dm *models.DiskModelImpl) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	warningStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11"))

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Select Target Disk"))
	b.WriteString("\n\n")

	if dm.Error() != nil {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("Error detecting disks: " + dm.Error().Error()))
		b.WriteString("\n")
	}

	if !dm.HasDisks() {
		b.WriteString(dimStyle.Render("No disks detected"))
		b.WriteString("\n")
	} else {
		for i, option := range dm.Options() {
			if i == dm.SelectedIndex() {
				b.WriteString(selectedStyle.Render("> " + option.Label))
			} else {
				b.WriteString(normalStyle.Render("  " + option.Label))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(warningStyle.Render("⚠ WARNING: Selected disk will be completely erased!"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("↑/↓ navigate • enter select • esc back • ctrl+c quit"))

	return b.String()
}
