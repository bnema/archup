package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/charmbracelet/lipgloss"
)

// RenderSummary renders the installation summary (success case)
func RenderSummary(im *models.InstallationModelImpl) string {
	if err := im.GetError(); err != "" {
		return RenderError(err)
	}

	_ = im.IsComplete() // keep for potential future use

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")).
		Render("Installation Complete!"))
	b.WriteString("\n\n")

	status := im.GetStatus()

	// System information
	if status.Hostname != "" {
		b.WriteString("Hostname: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(status.Hostname))
		b.WriteString("\n")
	}

	if status.Username != "" {
		b.WriteString("User: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(status.Username))
		b.WriteString("\n")
	}

	if status.TargetDisk != "" {
		b.WriteString("Target Disk: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(status.TargetDisk))
		b.WriteString("\n")
	}

	// Duration
	if status.StartedAt != nil && status.CompletedAt != nil {
		duration := status.CompletedAt.Sub(*status.StartedAt)
		b.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(duration)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Render("You can now reboot into your new system!"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Press 'q' to exit, then run: reboot"))

	return b.String()
}

// RenderError renders the error screen
func RenderError(errorMsg string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("1")).
		Render("Installation Failed"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Render(errorMsg))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Press 'q' or Ctrl+C to exit"))

	return b.String()
}

// RenderStatus renders the installation in-progress status
func RenderStatus(im *models.InstallationModelImpl) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Installation Status"))
	b.WriteString("\n\n")

	status := im.GetStatus()

	// Current phase
	if status.CurrentPhase != "" {
		b.WriteString("Phase: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(status.CurrentPhase))
		b.WriteString("\n")
	}

	// Progress
	if status.Progress > 0 {
		b.WriteString(fmt.Sprintf("Progress: %d%%\n", status.Progress))
	}

	// Status
	if status.State != "" {
		b.WriteString(fmt.Sprintf("State: %s\n", status.State))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Installation in progress... Press Ctrl+C to cancel"))

	return b.String()
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
