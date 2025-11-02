package phases

import (
	"fmt"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
)

// PhaseStatus represents the current status of a phase
type PhaseStatus int

const (
	StatusPending PhaseStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusSkipped
)

func (s PhaseStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// PhaseResult represents the result of a phase execution
type PhaseResult struct {
	Success bool
	Message string
	Error   error
}

// ProgressUpdate represents a progress update from a phase
type ProgressUpdate struct {
	PhaseName   string
	Step        string
	OutputLine  string
	Current     int
	Total       int
	IsComplete  bool
	IsError     bool
	ErrorMsg    string
}

// Phase represents a single installation phase
type Phase interface {
	// Name returns the phase identifier
	Name() string

	// Description returns a human-readable description
	Description() string

	// PreCheck validates if the phase can run
	// Returns nil if ready, error if preconditions not met
	PreCheck() error

	// Execute runs the phase logic
	// Should send ProgressUpdate via progressChan
	Execute(progressChan chan<- ProgressUpdate) PhaseResult

	// PostCheck validates the phase completed successfully
	// Returns nil if validation passes, error otherwise
	PostCheck() error

	// Rollback attempts to undo phase changes on failure
	// Returns nil if rollback successful, error otherwise
	Rollback() error

	// CanSkip returns true if this phase can be skipped
	CanSkip() bool

	// Status returns the current phase status
	Status() PhaseStatus

	// SetStatus updates the phase status
	SetStatus(status PhaseStatus)
}

// BasePhase provides common functionality for all phases
type BasePhase struct {
	name        string
	description string
	status      PhaseStatus
	config      *config.Config
	logger      *logger.Logger
}

// NewBasePhase creates a new base phase
func NewBasePhase(name, description string, cfg *config.Config, log *logger.Logger) *BasePhase {
	return &BasePhase{
		name:        name,
		description: description,
		status:      StatusPending,
		config:      cfg,
		logger:      log,
	}
}

// Name returns the phase name
func (b *BasePhase) Name() string {
	return b.name
}

// Description returns the phase description
func (b *BasePhase) Description() string {
	return b.description
}

// Status returns the current status
func (b *BasePhase) Status() PhaseStatus {
	return b.status
}

// SetStatus updates the status
func (b *BasePhase) SetStatus(status PhaseStatus) {
	b.status = status
}

// Config returns the configuration
func (b *BasePhase) Config() *config.Config {
	return b.config
}

// Logger returns the logger
func (b *BasePhase) Logger() *logger.Logger {
	return b.logger
}

// SendProgress is a helper to send progress updates
func (b *BasePhase) SendProgress(progressChan chan<- ProgressUpdate, step string, current, total int) {
	if progressChan != nil {
		progressChan <- ProgressUpdate{
			PhaseName:  b.name,
			Step:       step,
			Current:    current,
			Total:      total,
			IsComplete: false,
			IsError:    false,
		}
	}
}

// SendComplete is a helper to send completion
func (b *BasePhase) SendComplete(progressChan chan<- ProgressUpdate, message string) {
	if progressChan != nil {
		progressChan <- ProgressUpdate{
			PhaseName:  b.name,
			Step:       message,
			IsComplete: true,
			IsError:    false,
		}
	}
}

// SendError is a helper to send error
func (b *BasePhase) SendError(progressChan chan<- ProgressUpdate, err error) {
	if progressChan != nil {
		progressChan <- ProgressUpdate{
			PhaseName:  b.name,
			IsComplete: true,
			IsError:    true,
			ErrorMsg:   err.Error(),
		}
	}
}

// SendOutput is a helper to send command output line
func (b *BasePhase) SendOutput(progressChan chan<- ProgressUpdate, line string) {
	// Log the output
	if b.logger != nil {
		b.logger.Info(line)
	}

	// Send to progress channel for TUI display
	if progressChan != nil {
		progressChan <- ProgressUpdate{
			PhaseName:  b.name,
			OutputLine: line,
		}
	}
}

// DefaultPreCheck provides a basic pre-check (override in specific phases)
func (b *BasePhase) PreCheck() error {
	if b.config == nil {
		return fmt.Errorf("config not initialized")
	}
	return nil
}

// DefaultPostCheck provides a basic post-check (override in specific phases)
func (b *BasePhase) PostCheck() error {
	return nil
}

// DefaultRollback provides a basic rollback (override in specific phases)
func (b *BasePhase) Rollback() error {
	return fmt.Errorf("rollback not implemented for phase: %s", b.name)
}

// DefaultCanSkip returns false by default (override in specific phases)
func (b *BasePhase) CanSkip() bool {
	return false
}
