package views

import (
	"fmt"
	"strings"

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

// Render renders the logo
func (l *Logo) Render() string {
	lines := strings.Split(logoASCII, "\n")
	var rendered strings.Builder

	for _, line := range lines {
		if len(line) == 0 {
			rendered.WriteString("\n")
			continue
		}

		rendered.WriteString(line)
		rendered.WriteString("\n")
	}

	rendered.WriteString("\n")
	rendered.WriteString(fmt.Sprintf("  %s\n", l.version))

	return rendered.String()
}

// CenterText centers text within a given width using lipgloss
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(text)
}
