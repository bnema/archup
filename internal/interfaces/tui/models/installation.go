package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/bnema/archup/internal/application/dto"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallationModelImpl implements InstallationModel interface
type InstallationModelImpl struct {
	BaseModel

	status   *dto.InstallationStatus
	err      string
	notice   string
	complete bool
}

// NewInstallationModel creates a new installation model
func NewInstallationModel() *InstallationModelImpl {
	return &InstallationModelImpl{
		status:   &dto.InstallationStatus{},
		err:      "",
		notice:   "",
		complete: false,
	}
}

// Init initializes the model
func (im *InstallationModelImpl) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (im *InstallationModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return im, tea.Quit
		}
	}
	return im, nil
}

// View renders the installation status or summary
// Deprecated: Use views.RenderSummary instead
func (im *InstallationModelImpl) View() string {
	if im.err != "" {
		return im.ViewError()
	}

	if im.complete {
		return im.renderSummary()
	}

	return im.renderStatus()
}

// ViewError renders the error screen
// Deprecated: Use views.RenderError instead
func (im *InstallationModelImpl) ViewError() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("1")).
		Render("Installation Failed"))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Render(im.err))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Press 'q' or Ctrl+C to exit"))

	return b.String()
}

// SetError sets an error message
func (im *InstallationModelImpl) SetError(message string) {
	im.err = message
}

func (im *InstallationModelImpl) SetNotice(message string) {
	im.notice = message
}

// SetComplete marks installation as complete
func (im *InstallationModelImpl) SetComplete() {
	im.complete = true
}

// SetStatus sets the installation status
func (im *InstallationModelImpl) SetStatus(status *dto.InstallationStatus) {
	im.status = status
}

// GetError returns the error message if any
func (im *InstallationModelImpl) GetError() string {
	return im.err
}

func (im *InstallationModelImpl) GetNotice() string {
	return im.notice
}

// IsComplete returns whether the installation is complete
func (im *InstallationModelImpl) IsComplete() bool {
	return im.complete
}

// GetStatus returns the installation status
func (im *InstallationModelImpl) GetStatus() *dto.InstallationStatus {
	return im.status
}

// Helper methods

func (im *InstallationModelImpl) renderStatus() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Installation Status"))
	b.WriteString("\n\n")

	// Current phase
	if im.status.CurrentPhase != "" {
		b.WriteString("Phase: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(im.status.CurrentPhase))
		b.WriteString("\n")
	}

	// Progress
	if im.status.Progress > 0 {
		b.WriteString(fmt.Sprintf("Progress: %d%%\n", im.status.Progress))
	}

	// Status
	if im.status.State != "" {
		b.WriteString(fmt.Sprintf("State: %s\n", im.status.State))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Installation in progress... Press Ctrl+C to cancel"))

	return b.String()
}

func (im *InstallationModelImpl) renderSummary() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10")).
		Render("Installation Complete!"))
	b.WriteString("\n\n")

	// System information
	if im.status.Hostname != "" {
		b.WriteString("Hostname: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(im.status.Hostname))
		b.WriteString("\n")
	}

	if im.status.Username != "" {
		b.WriteString("User: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(im.status.Username))
		b.WriteString("\n")
	}

	if im.status.TargetDisk != "" {
		b.WriteString("Target Disk: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render(im.status.TargetDisk))
		b.WriteString("\n")
	}

	// Duration
	if im.status.StartedAt != nil && im.status.CompletedAt != nil {
		duration := im.status.CompletedAt.Sub(*im.status.StartedAt)
		b.WriteString(fmt.Sprintf("Duration: %s\n", formatDuration(duration)))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Faint(true).
		Render("Press 'r' to unmount and reboot, or 'q' to exit"))

	return b.String()
}

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
