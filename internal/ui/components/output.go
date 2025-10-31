package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/bnema/archup/internal/ui/styles"
)

// OutputViewer displays scrolling command output (last N lines)
type OutputViewer struct {
	viewport viewport.Model
	lines    []string
	maxLines int
}

// NewOutputViewer creates a new output viewer
func NewOutputViewer(width, height, maxLines int) *OutputViewer {
	vp := viewport.New(width, height)
	vp.Style = styles.LogStyle

	return &OutputViewer{
		viewport: vp,
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

// AddLine adds a new line of output (keeps only last N lines)
func (ov *OutputViewer) AddLine(line string) {
	ov.lines = append(ov.lines, line)

	// Keep only last maxLines
	if len(ov.lines) > ov.maxLines {
		ov.lines = ov.lines[len(ov.lines)-ov.maxLines:]
	}

	// Update viewport content
	ov.viewport.SetContent(strings.Join(ov.lines, "\n"))
	// Auto-scroll to bottom
	ov.viewport.GotoBottom()
}

// Clear clears all output
func (ov *OutputViewer) Clear() {
	ov.lines = make([]string, 0, ov.maxLines)
	ov.viewport.SetContent("")
}

// Update handles viewport updates
func (ov *OutputViewer) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	ov.viewport, cmd = ov.viewport.Update(msg)
	return cmd
}

// View renders the viewport
func (ov *OutputViewer) View() string {
	return ov.viewport.View()
}

// SetSize updates the viewport dimensions
func (ov *OutputViewer) SetSize(width, height int) {
	ov.viewport.Width = width
	ov.viewport.Height = height
}
