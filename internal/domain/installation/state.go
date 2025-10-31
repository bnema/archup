package installation

// State represents the current phase of the installation
type State int

const (
	// StateNotStarted is the initial state before installation begins
	StateNotStarted State = iota

	// StatePreflightChecks represents preflight verification phase
	StatePreflightChecks

	// StateDiskPartitioning represents disk partitioning phase
	StateDiskPartitioning

	// StateBaseInstallation represents base system installation phase
	StateBaseInstallation

	// StateSystemConfiguration represents system configuration phase (hostname, locale, etc.)
	StateSystemConfiguration

	// StateBootloaderSetup represents bootloader installation phase
	StateBootloaderSetup

	// StateRepositorySetup represents repository configuration phase
	StateRepositorySetup

	// StatePostInstallation represents post-installation tasks
	StatePostInstallation

	// StateCompleted represents successful installation completion
	StateCompleted

	// StateFailed represents installation failure
	StateFailed
)

// String returns human-readable state name
func (s State) String() string {
	switch s {
	case StateNotStarted:
		return "NotStarted"
	case StatePreflightChecks:
		return "PreflightChecks"
	case StateDiskPartitioning:
		return "DiskPartitioning"
	case StateBaseInstallation:
		return "BaseInstallation"
	case StateSystemConfiguration:
		return "SystemConfiguration"
	case StateBootloaderSetup:
		return "BootloaderSetup"
	case StateRepositorySetup:
		return "RepositorySetup"
	case StatePostInstallation:
		return "PostInstallation"
	case StateCompleted:
		return "Completed"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// IsTerminal returns true if the state is a terminal state (cannot transition further)
func (s State) IsTerminal() bool {
	return s == StateCompleted || s == StateFailed
}

// IsValid returns true if the state is a valid state
func (s State) IsValid() bool {
	return s >= StateNotStarted && s <= StateFailed
}

// CanTransitionTo returns true if a transition from the current state to the target state is allowed
// This enforces the business rule that states must progress in strict order
func (s State) CanTransitionTo(target State) bool {
	// Cannot transition FROM a terminal state
	if s.IsTerminal() {
		return false
	}

	// Can always transition to Failed from any non-terminal state
	if target == StateFailed {
		return true
	}

	// Must be valid target
	if !target.IsValid() {
		return false
	}

	// Define valid transitions (strict progression)
	validTransitions := map[State][]State{
		StateNotStarted:          {StatePreflightChecks},
		StatePreflightChecks:     {StateDiskPartitioning},
		StateDiskPartitioning:    {StateBaseInstallation},
		StateBaseInstallation:    {StateSystemConfiguration},
		StateSystemConfiguration: {StateBootloaderSetup},
		StateBootloaderSetup:     {StateRepositorySetup},
		StateRepositorySetup:     {StatePostInstallation},
		StatePostInstallation:    {StateCompleted},
		StateCompleted:           {},
		StateFailed:              {},
	}

	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, allowedState := range allowed {
		if allowedState == target {
			return true
		}
	}

	return false
}

// TransitionTo performs a state transition with validation
func (s *State) TransitionTo(target State) error {
	if !s.CanTransitionTo(target) {
		return ErrInvalidStateTransition
	}
	*s = target
	return nil
}

// PhaseSequence returns the sequence of installation phases in order
func PhaseSequence() []State {
	return []State{
		StatePreflightChecks,
		StateDiskPartitioning,
		StateBaseInstallation,
		StateSystemConfiguration,
		StateBootloaderSetup,
		StateRepositorySetup,
		StatePostInstallation,
		StateCompleted,
	}
}

// PhaseIndex returns the index of a phase in the sequence, or -1 if not found
func PhaseIndex(state State) int {
	sequence := PhaseSequence()
	for i, s := range sequence {
		if s == state {
			return i
		}
	}
	return -1
}

// ProgressPercentage returns estimated progress as percentage (0-100) based on current state
func (s State) ProgressPercentage() int {
	switch s {
	case StateNotStarted:
		return 0
	case StatePreflightChecks:
		return 10
	case StateDiskPartitioning:
		return 25
	case StateBaseInstallation:
		return 40
	case StateSystemConfiguration:
		return 60
	case StateBootloaderSetup:
		return 75
	case StateRepositorySetup:
		return 85
	case StatePostInstallation:
		return 95
	case StateCompleted:
		return 100
	case StateFailed:
		return 0 // Reset to 0 on failure
	default:
		return 0
	}
}
