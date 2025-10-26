package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/bnema/archup/internal/ui/styles"
)

const (
	// Logo ASCII art (from logo.txt)
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
// version should be injected via ldflags: -ldflags "-X main.version=v1.0.0"
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

// Render renders the logo in full blue
func (l *Logo) Render() string {
	lines := strings.Split(logoASCII, "\n")
	var rendered strings.Builder

	for _, line := range lines {
		if len(line) == 0 {
			rendered.WriteString("\n")
			continue
		}

		// Render entire logo in blue
		rendered.WriteString(styles.LogoCyanStyle.Render(line))
		rendered.WriteString("\n")
	}

	// Add version below logo
	rendered.WriteString("\n")
	versionLine := "  " + l.version
	rendered.WriteString(styles.LogoCyanStyle.Render(versionLine))
	rendered.WriteString("\n")

	return rendered.String()
}

// RenderCentered renders the logo centered horizontally within the terminal
func (l *Logo) RenderCentered(termWidth int) string {
	logo := l.Render()
	lines := strings.Split(logo, "\n")

	var centered strings.Builder
	paddingLeft := (termWidth - l.width) / 2

	if paddingLeft < 0 {
		paddingLeft = 0
	}

	padding := strings.Repeat(" ", paddingLeft)

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			centered.WriteString(padding)
		}
		centered.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			centered.WriteString("\n")
		}
	}

	return centered.String()
}

// Width returns the logo width
func (l *Logo) Width() int {
	return l.width
}

// Height returns the logo height (including version line)
func (l *Logo) Height() int {
	return l.height + 2 // +2 for blank line and version
}

// ClearScreen clears the screen and displays the logo at the top
func ClearScreen(termWidth int, version string) string {
	logo := NewLogo(version)
	return fmt.Sprintf("\033[H\033[2J\n%s", logo.RenderCentered(termWidth))
}

// RenderWithMessage renders the logo with a message below it
func RenderWithMessage(termWidth int, version, message string) string {
	logo := NewLogo(version)
	logoStr := logo.RenderCentered(termWidth)

	// Center the message
	paddingLeft := (termWidth - lipgloss.Width(message)) / 2
	if paddingLeft < 0 {
		paddingLeft = 0
	}
	padding := strings.Repeat(" ", paddingLeft)

	return fmt.Sprintf("%s\n%s%s\n", logoStr, padding, message)
}
