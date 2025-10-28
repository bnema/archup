package views

import (
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

const (
	WelcomeTagline = "A minimal, slightly opinionated Arch Linux installer."

	logoASCII = `  ░█████╗░██████╗░░█████╗░██╗░░██╗██╗░░░██╗██████╗░
  ██╔══██╗██╔══██╗██╔══██╗██║░░██║██║░░░██║██╔══██╗
  ███████║██████╔╝██║░░╚═╝███████║██║░░░██║██████╔╝
  ██╔══██║██╔══██╗██║░░██╗██╔══██║██║░░░██║██╔═══╝░
  ██║░░██║██║░░██║╚█████╔╝██║░░██║╚██████╔╝██║░░░░░
  ╚═╝░░╚═╝╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝░╚═════╝░╚═╝░░░░░`
)

// Logo represents the ArchUp logo display
type Logo struct {
	width   int
	height  int
	version string
}

// NewLogo creates a new Logo instance
func NewLogo(version string) *Logo {
	lines := strings.Split(logoASCII, "\n")
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	if version == "" {
		version = "dev"
	}

	return &Logo{
		width:   width,
		height:  len(lines),
		version: version,
	}
}

// RenderCentered renders the logo centered using lipgloss alignment
func (l *Logo) RenderCentered(termWidth int) string {
	return l.Render()
}

// Render renders the logo with theme colors
func (l *Logo) Render() string {
	lines := strings.Split(logoASCII, "\n")
	var rendered strings.Builder

	for _, line := range lines {
		if len(line) == 0 {
			rendered.WriteString("\n")
			continue
		}

		// Apply bright cyan color to the logo
		rendered.WriteString(styles.LogoCyanStyle.Render(line))
		rendered.WriteString("\n")
	}

	rendered.WriteString("\n")
	// Apply dimmed style to version
	rendered.WriteString(styles.HelpStyle.Render(fmt.Sprintf("  %s\n", l.version)))

	return rendered.String()
}

// CenterText centers text within a given width using lipgloss
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(text)
}
