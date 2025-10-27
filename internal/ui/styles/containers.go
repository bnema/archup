package styles

import "github.com/charmbracelet/lipgloss"

// Container styles for consistent layout
var (
	// CenteredContainerStyle - Used for welcome, complete, error screens
	CenteredContainerStyle = lipgloss.NewStyle().
		MaxWidth(80).
		Padding(0, 2).
		AlignHorizontal(lipgloss.Center)

	// FormContainerStyle - Used for all forms
	FormContainerStyle = lipgloss.NewStyle().
		MaxWidth(80).
		Padding(0, 2).
		AlignHorizontal(lipgloss.Left)
)
