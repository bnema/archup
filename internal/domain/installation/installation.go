package installation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Installation is the aggregate root entity representing the entire installation process
// It manages the lifecycle of an Arch Linux installation from start to completion
type Installation struct {
	// Unique identifier for this installation
	id string

	// Current state of the installation
	state State

	// Configuration for the installation (hostname, username, etc.)
	// These are immutable once set
	hostname       string
	username       string
	targetDisk     string
	encryptionType string

	// Installation metadata
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time

	// Domain events that have occurred during this installation
	events []DomainEvent
}

// NewInstallation creates a new Installation aggregate
// All parameters are validated according to business rules
func NewInstallation(
	hostname string,
	username string,
	targetDisk string,
	encryptionType string,
) (*Installation, error) {
	// Validate hostname
	if hostname == "" {
		return nil, errors.New("hostname cannot be empty")
	}
	if len(hostname) > 63 {
		return nil, errors.New("hostname cannot exceed 63 characters")
	}

	// Validate username
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if len(username) > 32 {
		return nil, errors.New("username cannot exceed 32 characters")
	}

	// Validate targetDisk
	if targetDisk == "" {
		return nil, errors.New("target disk cannot be empty")
	}

	// Validate encryptionType
	validEncryption := map[string]bool{
		"none":     true,
		"luks":     true,
		"luks-lvm": true,
	}
	if !validEncryption[encryptionType] {
		return nil, fmt.Errorf("invalid encryption type: %s", encryptionType)
	}

	installation := &Installation{
		id:             uuid.New().String(),
		state:          StateNotStarted,
		hostname:       hostname,
		username:       username,
		targetDisk:     targetDisk,
		encryptionType: encryptionType,
		createdAt:      time.Now().UTC(),
		events:         []DomainEvent{},
	}

	return installation, nil
}

// ID returns the unique identifier
func (i *Installation) ID() string {
	return i.id
}

// State returns the current installation state
func (i *Installation) State() State {
	return i.state
}

// Hostname returns the configured hostname
func (i *Installation) Hostname() string {
	return i.hostname
}

// Username returns the configured username
func (i *Installation) Username() string {
	return i.username
}

// TargetDisk returns the target installation disk
func (i *Installation) TargetDisk() string {
	return i.targetDisk
}

// EncryptionType returns the encryption type ("none", "luks", or "luks-lvm")
func (i *Installation) EncryptionType() string {
	return i.encryptionType
}

// CreatedAt returns when this installation was created
func (i *Installation) CreatedAt() time.Time {
	return i.createdAt
}

// StartedAt returns when the installation actually started (or nil if not started)
func (i *Installation) StartedAt() *time.Time {
	return i.startedAt
}

// CompletedAt returns when the installation completed (or nil if not completed)
func (i *Installation) CompletedAt() *time.Time {
	return i.completedAt
}

// IsStarted returns true if the installation has been started
func (i *Installation) IsStarted() bool {
	return i.state != StateNotStarted
}

// IsCompleted returns true if the installation has completed (successfully or failed)
func (i *Installation) IsCompleted() bool {
	return i.state.IsTerminal()
}

// IsSuccessful returns true if the installation completed successfully
func (i *Installation) IsSuccessful() bool {
	return i.state == StateCompleted
}

// IsFailed returns true if the installation failed
func (i *Installation) IsFailed() bool {
	return i.state == StateFailed
}

// Start transitions the installation from NotStarted to PreflightChecks
// This is the entry point that validates prerequisites and begins the process
func (i *Installation) Start(ctx context.Context) error {
	if i.state != StateNotStarted {
		return ErrInstallationAlreadyStarted
	}

	now := time.Now().UTC()
	i.startedAt = &now

	// Transition to preflight checks
	if err := i.state.TransitionTo(StatePreflightChecks); err != nil {
		return err
	}
	i.state = StatePreflightChecks

	// Emit event
	event := NewInstallationStartedEvent(
		i.id,
		i.hostname,
		i.username,
		i.targetDisk,
		i.encryptionType,
	)
	i.recordEvent(event)

	return nil
}

// TransitionToNextPhase moves the installation to the next phase in the sequence
func (i *Installation) TransitionToNextPhase() error {
	if i.state.IsTerminal() {
		return ErrInstallationAlreadyCompleted
	}

	nextState := i.getNextPhase()
	if nextState == -1 {
		return ErrInvalidPhaseTransition
	}

	if err := i.state.TransitionTo(State(nextState)); err != nil {
		return err
	}
	i.state = State(nextState)

	// Emit event
	event := NewPhaseStartedEvent(i.id, i.state)
	i.recordEvent(event)

	return nil
}

// CompleteCurrentPhase marks the current phase as completed and transitions to next
func (i *Installation) CompleteCurrentPhase(durationSeconds int) error {
	if i.state == StateNotStarted || i.state.IsTerminal() {
		return ErrInstallationNotStarted
	}

	currentPhase := i.state

	// Record completion event
	nextPhase := i.getNextPhase()
	event := NewPhaseCompletedEvent(i.id, currentPhase, State(nextPhase), durationSeconds)
	i.recordEvent(event)

	// Transition to next phase
	return i.TransitionToNextPhase()
}

// FailCurrentPhase marks the current phase as failed
func (i *Installation) FailCurrentPhase(errorMessage string, recoverable bool) error {
	if i.state == StateNotStarted {
		return ErrInstallationNotStarted
	}

	if i.state.IsTerminal() {
		return ErrInstallationAlreadyCompleted
	}

	currentPhase := i.state

	// Record failure event
	event := NewPhaseFailedEvent(i.id, currentPhase, errorMessage, recoverable)
	i.recordEvent(event)

	// Transition to failed state
	if err := i.state.TransitionTo(StateFailed); err != nil {
		return err
	}
	i.state = StateFailed

	// Record installation failure event
	failureEvent := NewInstallationFailedEvent(i.id, currentPhase, errorMessage)
	i.recordEvent(failureEvent)

	return nil
}

// Complete marks the installation as successfully completed
// Can be called from any non-terminal state (phases are tracked by progress tracker, not state machine)
func (i *Installation) Complete(totalDurationSeconds int) error {
	if i.state.IsTerminal() {
		return ErrInstallationAlreadyCompleted
	}
	if i.state == StateNotStarted {
		return ErrInstallationNotStarted
	}

	now := time.Now().UTC()
	i.completedAt = &now
	i.state = StateCompleted

	// Emit completion event
	event := NewInstallationCompletedEvent(i.id, totalDurationSeconds)
	i.recordEvent(event)

	return nil
}

// ProgressPercentage returns estimated progress as percentage (0-100)
func (i *Installation) ProgressPercentage() int {
	return i.state.ProgressPercentage()
}

// UncommittedEvents returns all domain events that have occurred but not yet persisted
func (i *Installation) UncommittedEvents() []DomainEvent {
	return i.events
}

// ClearEvents clears the uncommitted events after persistence
func (i *Installation) ClearEvents() {
	i.events = []DomainEvent{}
}

// Private methods

// recordEvent adds a domain event to the uncommitted events list
func (i *Installation) recordEvent(event DomainEvent) {
	i.events = append(i.events, event)
}

// getNextPhase returns the next phase in the sequence
func (i *Installation) getNextPhase() State {
	sequence := PhaseSequence()
	currentIndex := PhaseIndex(i.state)

	if currentIndex == -1 || currentIndex >= len(sequence)-1 {
		return -1
	}

	return sequence[currentIndex+1]
}
