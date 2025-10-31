package installation

import "time"

// DomainEvent is the base interface for all domain events
type DomainEvent interface {
	// AggregateID returns the installation ID this event relates to
	AggregateID() string

	// EventType returns the type of event
	EventType() string

	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
}

// BaseDomainEvent provides common fields for all domain events
type BaseDomainEvent struct {
	aggregateID string
	occurredAt  time.Time
}

func NewBaseDomainEvent(aggregateID string) BaseDomainEvent {
	return BaseDomainEvent{
		aggregateID: aggregateID,
		occurredAt:  time.Now().UTC(),
	}
}

func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// InstallationStartedEvent represents the start of an installation
type InstallationStartedEvent struct {
	BaseDomainEvent
	Hostname       string
	Username       string
	TargetDisk     string
	EncryptionType string
}

func NewInstallationStartedEvent(
	aggregateID string,
	hostname string,
	username string,
	targetDisk string,
	encryptionType string,
) *InstallationStartedEvent {
	return &InstallationStartedEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		Hostname:        hostname,
		Username:        username,
		TargetDisk:      targetDisk,
		EncryptionType:  encryptionType,
	}
}

func (e *InstallationStartedEvent) EventType() string {
	return "InstallationStarted"
}

// PhaseStartedEvent represents the start of a specific phase
type PhaseStartedEvent struct {
	BaseDomainEvent
	Phase State
}

func NewPhaseStartedEvent(aggregateID string, phase State) *PhaseStartedEvent {
	return &PhaseStartedEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		Phase:           phase,
	}
}

func (e *PhaseStartedEvent) EventType() string {
	return "PhaseStarted"
}

// PhaseCompletedEvent represents successful completion of a phase
type PhaseCompletedEvent struct {
	BaseDomainEvent
	Phase           State
	PreviousState   State
	DurationSeconds int
}

func NewPhaseCompletedEvent(aggregateID string, phase State, previousState State, durationSeconds int) *PhaseCompletedEvent {
	return &PhaseCompletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		Phase:           phase,
		PreviousState:   previousState,
		DurationSeconds: durationSeconds,
	}
}

func (e *PhaseCompletedEvent) EventType() string {
	return "PhaseCompleted"
}

// PhaseFailedEvent represents failure of a phase
type PhaseFailedEvent struct {
	BaseDomainEvent
	Phase       State
	ErrorMessage string
	Recoverable bool
}

func NewPhaseFailedEvent(aggregateID string, phase State, errorMessage string, recoverable bool) *PhaseFailedEvent {
	return &PhaseFailedEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		Phase:           phase,
		ErrorMessage:    errorMessage,
		Recoverable:     recoverable,
	}
}

func (e *PhaseFailedEvent) EventType() string {
	return "PhaseFailed"
}

// InstallationCompletedEvent represents successful installation completion
type InstallationCompletedEvent struct {
	BaseDomainEvent
	TotalDurationSeconds int
}

func NewInstallationCompletedEvent(aggregateID string, totalDurationSeconds int) *InstallationCompletedEvent {
	return &InstallationCompletedEvent{
		BaseDomainEvent:      NewBaseDomainEvent(aggregateID),
		TotalDurationSeconds: totalDurationSeconds,
	}
}

func (e *InstallationCompletedEvent) EventType() string {
	return "InstallationCompleted"
}

// InstallationFailedEvent represents installation failure
type InstallationFailedEvent struct {
	BaseDomainEvent
	FailedPhase  State
	ErrorMessage string
}

func NewInstallationFailedEvent(aggregateID string, failedPhase State, errorMessage string) *InstallationFailedEvent {
	return &InstallationFailedEvent{
		BaseDomainEvent: NewBaseDomainEvent(aggregateID),
		FailedPhase:     failedPhase,
		ErrorMessage:    errorMessage,
	}
}

func (e *InstallationFailedEvent) EventType() string {
	return "InstallationFailed"
}
