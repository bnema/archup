package phases

import (
	"fmt"

	"github.com/bnema/archup/internal/config"
)

// Orchestrator manages the execution of installation phases
type Orchestrator struct {
	phases      []Phase
	currentIdx  int
	config      *config.Config
	logPath     string
	progressCh  chan ProgressUpdate
}

// NewOrchestrator creates a new phase orchestrator
func NewOrchestrator(cfg *config.Config, logPath string) *Orchestrator {
	return &Orchestrator{
		phases:     make([]Phase, 0),
		currentIdx: -1,
		config:     cfg,
		logPath:    logPath,
		progressCh: make(chan ProgressUpdate, 100),
	}
}

// RegisterPhase adds a phase to the execution sequence
func (o *Orchestrator) RegisterPhase(phase Phase) error {
	if phase == nil {
		return fmt.Errorf("cannot register nil phase")
	}

	// Check for duplicate phase names
	for _, p := range o.phases {
		if p.Name() == phase.Name() {
			return fmt.Errorf("phase %s already registered", phase.Name())
		}
	}

	o.phases = append(o.phases, phase)
	return nil
}

// Phases returns all registered phases
func (o *Orchestrator) Phases() []Phase {
	return o.phases
}

// CurrentPhase returns the currently executing phase, or nil
func (o *Orchestrator) CurrentPhase() Phase {
	if o.currentIdx < 0 || o.currentIdx >= len(o.phases) {
		return nil
	}
	return o.phases[o.currentIdx]
}

// NextPhase returns the next phase to execute, or nil if done
func (o *Orchestrator) NextPhase() Phase {
	// Find next pending phase
	for i := o.currentIdx + 1; i < len(o.phases); i++ {
		phase := o.phases[i]
		if phase.Status() == StatusPending {
			return phase
		}
	}
	return nil
}

// ProgressChannel returns the progress update channel
func (o *Orchestrator) ProgressChannel() <-chan ProgressUpdate {
	return o.progressCh
}

// ExecutePhase runs a specific phase
func (o *Orchestrator) ExecutePhase(phase Phase) error {
	// Update current index
	for i, p := range o.phases {
		if p.Name() == phase.Name() {
			o.currentIdx = i
			break
		}
	}

	// Pre-check
	phase.SetStatus(StatusRunning)
	if err := phase.PreCheck(); err != nil {
		phase.SetStatus(StatusFailed)
		return fmt.Errorf("pre-check failed: %w", err)
	}

	// Execute
	result := phase.Execute(o.progressCh)
	if !result.Success {
		phase.SetStatus(StatusFailed)

		// Attempt rollback
		if rbErr := phase.Rollback(); rbErr != nil {
			return fmt.Errorf("execution failed: %v, rollback failed: %v", result.Error, rbErr)
		}

		return fmt.Errorf("execution failed: %w", result.Error)
	}

	// Post-check
	if err := phase.PostCheck(); err != nil {
		phase.SetStatus(StatusFailed)
		return fmt.Errorf("post-check failed: %w", err)
	}

	phase.SetStatus(StatusCompleted)
	return nil
}

// ExecuteNext executes the next pending phase
func (o *Orchestrator) ExecuteNext() error {
	phase := o.NextPhase()
	if phase == nil {
		return fmt.Errorf("no pending phases")
	}
	return o.ExecutePhase(phase)
}

// ExecuteAll executes all pending phases in sequence
func (o *Orchestrator) ExecuteAll() error {
	for {
		phase := o.NextPhase()
		if phase == nil {
			break
		}

		if err := o.ExecutePhase(phase); err != nil {
			return err
		}
	}
	return nil
}

// SkipPhase marks a phase as skipped
func (o *Orchestrator) SkipPhase(phase Phase) error {
	if !phase.CanSkip() {
		return fmt.Errorf("phase %s cannot be skipped", phase.Name())
	}
	phase.SetStatus(StatusSkipped)
	return nil
}

// Reset resets all phases to pending state
func (o *Orchestrator) Reset() {
	o.currentIdx = -1
	for _, phase := range o.phases {
		phase.SetStatus(StatusPending)
	}
}

// Progress returns overall installation progress
func (o *Orchestrator) Progress() (completed int, total int) {
	total = len(o.phases)
	for _, phase := range o.phases {
		if phase.Status() == StatusCompleted || phase.Status() == StatusSkipped {
			completed++
		}
	}
	return completed, total
}

// IsComplete returns true if all phases are done
func (o *Orchestrator) IsComplete() bool {
	for _, phase := range o.phases {
		status := phase.Status()
		if status != StatusCompleted && status != StatusSkipped {
			return false
		}
	}
	return true
}

// HasFailed returns true if any phase has failed
func (o *Orchestrator) HasFailed() bool {
	for _, phase := range o.phases {
		if phase.Status() == StatusFailed {
			return true
		}
	}
	return false
}
