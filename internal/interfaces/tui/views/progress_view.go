package views

import (
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderProgress renders the progress model to a styled string
func RenderProgress(pm *models.ProgressModelImpl) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Installation Progress"))
	b.WriteString("\n\n")

	// Phase indicator
	phaseNumber, totalPhases := pm.GetPhaseProgress()
	if totalPhases > 0 {
		phasePercent := int((float64(phaseNumber) / float64(totalPhases)) * 100)
		currentPhase := pm.GetCurrentPhase()
		b.WriteString(fmt.Sprintf("Phase %d/%d: %s [%d%%]\n", phaseNumber, totalPhases, currentPhase, phasePercent))
	}

	// Progress bar
	b.WriteString("\n")
	b.WriteString(renderProgressBar(pm))
	b.WriteString("\n\n")

	// Current message
	message := pm.GetMessage()
	if message != "" {
		color := lipgloss.Color("11")
		if pm.IsError() {
			color = lipgloss.Color("1")
		}
		b.WriteString("Current: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(color).
			Render(message))
		b.WriteString("\n\n")
	}

	// Message history
	history := pm.GetMessageHistory()
	if len(history) > 0 {
		b.WriteString(lipgloss.NewStyle().
			Faint(true).
			Render("Recent activities:"))
		b.WriteString("\n")

		start := 0
		if len(history) > 5 {
			start = len(history) - 5
		}
		for _, msg := range history[start:] {
			b.WriteString(lipgloss.NewStyle().
				Faint(true).
				Render("  • " + msg))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Installation running... Press Ctrl+C to cancel"))

	return b.String()
}

// renderProgressBar renders a visual progress bar
func renderProgressBar(pm *models.ProgressModelImpl) string {
	width := 40
	progressPercent := pm.GetProgressPercent()
	filled := (progressPercent * width) / 100

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("[%s] %d%%", bar, progressPercent)
}
