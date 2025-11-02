package styles

import "github.com/charmbracelet/lipgloss"

// Layout constants
const (
	MaxFormWidth    = 80  // Maximum width for form containers
	MaxContentWidth = 120 // Maximum width for content (welcome, complete screens)
)

// Alignment aliases for clarity
const (
	LeftAlign   = lipgloss.Left
	CenterAlign = lipgloss.Center
)

// Container styles for consistent layout
var (
	// CenteredContainerStyle - Used for welcome, complete, error screens
	CenteredContainerStyle = lipgloss.NewStyle().
		MaxWidth(MaxContentWidth).
		Padding(0, 2).
		AlignHorizontal(lipgloss.Center)

	// FormContainerStyle - Used for all forms (content left-aligned)
	FormContainerStyle = lipgloss.NewStyle().
		Padding(0, 2).
		AlignHorizontal(lipgloss.Left)
)
