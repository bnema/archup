package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bnema/archup/internal/application/dto"
)

// ProgressModelImpl implements ProgressModel interface
type ProgressModelImpl struct {
	BaseModel

	currentPhase     string
	phaseNumber      int
	totalPhases      int
	progressPercent  int
	message          string
	isError          bool
	lastMessage      string
	messageHistory   []string
	maxHistoryLength int
}

// NewProgressModel creates a new progress model
func NewProgressModel() *ProgressModelImpl {
	return &ProgressModelImpl{
		maxHistoryLength: 10,
		messageHistory:   []string{},
	}
}

// Init initializes the model
func (pm *ProgressModelImpl) Init() tea.Cmd {
	return nil
}

// UpdateProgress handles progress updates (called from app when progress updates arrive)
func (pm *ProgressModelImpl) UpdateProgress(update *dto.ProgressUpdate) {
	if update == nil {
		return
	}

	pm.currentPhase = update.Phase
	pm.phaseNumber = update.PhaseNumber
	pm.totalPhases = update.TotalPhases
	pm.progressPercent = update.ProgressPercent
	pm.message = update.Message
	pm.isError = update.IsError

	// Add to history
	if pm.message != "" && pm.message != pm.lastMessage {
		pm.addToHistory(pm.message)
		pm.lastMessage = pm.message
	}
}

// Update handles tea.Msg (for tea.Model interface compatibility)
func (pm *ProgressModelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return pm, nil
}

// View renders the progress display
// Deprecated: Use views.RenderProgress instead
func (pm *ProgressModelImpl) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Installation Progress"))
	b.WriteString("\n\n")

	// Phase indicator
	if pm.totalPhases > 0 {
		phasePercent := int((float64(pm.phaseNumber) / float64(pm.totalPhases)) * 100)
		b.WriteString(fmt.Sprintf("Phase %d/%d: %s [%d%%]\n", pm.phaseNumber, pm.totalPhases, pm.currentPhase, phasePercent))
	}

	// Progress bar
	b.WriteString("\n")
	b.WriteString(pm.renderProgressBar())
	b.WriteString("\n\n")

	// Current message
	if pm.message != "" {
		color := lipgloss.Color("11")
		if pm.isError {
			color = lipgloss.Color("1")
		}
		b.WriteString("Current: ")
		b.WriteString(lipgloss.NewStyle().
			Foreground(color).
			Render(pm.message))
		b.WriteString("\n\n")
	}

	// Message history
	if len(pm.messageHistory) > 0 {
		b.WriteString(lipgloss.NewStyle().
			Faint(true).
			Render("Recent activities:"))
		b.WriteString("\n")

		start := 0
		if len(pm.messageHistory) > 5 {
			start = len(pm.messageHistory) - 5
		}
		for _, msg := range pm.messageHistory[start:] {
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

// GetPhaseProgress returns current and total phases
func (pm *ProgressModelImpl) GetPhaseProgress() (int, int) {
	return pm.phaseNumber, pm.totalPhases
}

// GetProgressPercent returns the overall progress percentage
func (pm *ProgressModelImpl) GetProgressPercent() int {
	return pm.progressPercent
}

// GetCurrentPhase returns the current phase name
func (pm *ProgressModelImpl) GetCurrentPhase() string {
	return pm.currentPhase
}

// GetMessage returns the current message
func (pm *ProgressModelImpl) GetMessage() string {
	return pm.message
}

// IsError returns whether the current message is an error
func (pm *ProgressModelImpl) IsError() bool {
	return pm.isError
}

// GetMessageHistory returns the message history
func (pm *ProgressModelImpl) GetMessageHistory() []string {
	return pm.messageHistory
}

// Helper methods

func (pm *ProgressModelImpl) renderProgressBar() string {
	width := 40
	filled := (pm.progressPercent * width) / 100

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("[%s] %d%%", bar, pm.progressPercent)
}

func (pm *ProgressModelImpl) addToHistory(msg string) {
	pm.messageHistory = append(pm.messageHistory, msg)
	if len(pm.messageHistory) > pm.maxHistoryLength {
		pm.messageHistory = pm.messageHistory[1:]
	}
}
