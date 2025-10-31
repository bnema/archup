package views

import (
	"strings"

	"github.com/bnema/archup/internal/ui/assets"
	"github.com/bnema/archup/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

const (
	WelcomeTagline = "A minimal, slightly opinionated Arch Linux installer."
)

// Logo represents the ArchUp logo display
type Logo struct {
	width   int
	height  int
	version string
}

// NewLogo creates a new Logo instance
func NewLogo(version string) *Logo {
	lines := strings.Split(assets.LogoASCII, "\n")
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

// Render renders the logo with character-level styling
func (l *Logo) Render() string {
	lines := strings.Split(assets.LogoASCII, "\n")
	var rendered strings.Builder

	for _, line := range lines {
		if len(line) == 0 {
			rendered.WriteString("\n")
			continue
		}

		// Apply conditional styling based on character
		var styledLine strings.Builder
		for _, ch := range line {
			if string(ch) == "░" {
				// Use darker blue for border character
				styledLine.WriteString(styles.LogoBlueStyle.Render(string(ch)))
			} else {
				// Use bright cyan for main logo
				styledLine.WriteString(styles.LogoCyanStyle.Render(string(ch)))
			}
		}
		rendered.WriteString(styledLine.String())
		rendered.WriteString("\n")
	}

	// Add version below logo
	rendered.WriteString("\n")
	versionLine := "  " + l.version
	rendered.WriteString(styles.HelpStyle.Render(versionLine))
	rendered.WriteString("\n")

	return rendered.String()
}

// CenterText centers text within a given width using lipgloss
func CenterText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(text)
}
