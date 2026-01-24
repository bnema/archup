package tui

import (
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/interfaces/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

// FormModel manages form state for user input
type FormModel interface {
	tea.Model
	GetData() models.FormData
	SetData(data models.FormData)
}

// InstallationModel manages installation state and summary
type InstallationModel interface {
	tea.Model
	View() string
	ViewError() string
	SetError(message string)
	SetComplete()
	SetStatus(status *dto.InstallationStatus)
	GetError() string
	IsComplete() bool
	GetStatus() *dto.InstallationStatus
}

// ProgressModel manages progress display state
type ProgressModel interface {
	tea.Model
	UpdateProgress(update *dto.ProgressUpdate)
	GetPhaseProgress() (current int, total int)
	GetProgressPercent() int
	GetCurrentPhase() string
	GetMessage() string
	IsError() bool
	GetMessageHistory() []string
}

// Type aliases for convenience
type FormData = models.FormData

// BaseModel provides common functionality for models
type BaseModel struct {
	width  int
	height int
}

// SetSize sets the terminal size for the model
func (bm *BaseModel) SetSize(width, height int) {
	bm.width = width
	bm.height = height
}

// GetSize returns the terminal size
func (bm *BaseModel) GetSize() (int, int) {
	return bm.width, bm.height
}
