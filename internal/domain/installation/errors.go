package installation

import "errors"

// Domain-specific error types for Installation aggregate

var (
	// ErrInvalidStateTransition occurs when attempting an invalid state change
	ErrInvalidStateTransition = errors.New("invalid state transition")

	// ErrInstallationAlreadyStarted occurs when trying to start an already started installation
	ErrInstallationAlreadyStarted = errors.New("installation already started")

	// ErrInstallationNotStarted occurs when operating on a non-started installation
	ErrInstallationNotStarted = errors.New("installation not started")

	// ErrInstallationAlreadyCompleted occurs when trying to modify a completed installation
	ErrInstallationAlreadyCompleted = errors.New("installation already completed")

	// ErrInvalidPhaseTransition occurs when phase transition violates business rules
	ErrInvalidPhaseTransition = errors.New("invalid phase transition")

	// ErrEncryptionPasswordRequired occurs when encryption chosen but no password provided
	ErrEncryptionPasswordRequired = errors.New("encryption password required for LUKS encryption")

	// ErrRootPartitionTooSmall occurs when root partition is smaller than minimum size
	ErrRootPartitionTooSmall = errors.New("root partition must be at least 20GB")

	// ErrBootPartitionTooSmall occurs when boot partition is smaller than minimum size
	ErrBootPartitionTooSmall = errors.New("boot partition must be at least 300MB")
)
