package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderCompositor renders compositor selection.
func RenderCompositor(config domain.DesktopConfig, selected int) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	options := []domain.Compositor{domain.CompositorNiri, domain.CompositorHyprland}

	b.WriteString("\n")
	b.WriteString(title.Render("Select Compositor"))
	b.WriteString("\n\n")

	for i, option := range options {
		label := string(option)
		if i == selected {
			b.WriteString(active.Render("› " + label))
		} else {
			b.WriteString("  " + label)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(muted.Render("↑↓ Move • Enter Select • Ctrl+C Quit"))

	_ = config
	return b.String()
}
