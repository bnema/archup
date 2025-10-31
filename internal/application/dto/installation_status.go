package dto

import "time"

// InstallationStatus represents the overall installation status
type InstallationStatus struct {
	ID                 string    // Installation ID (UUID)
	State              string    // Current state (NotStarted, PreflightChecks, etc.)
	Hostname           string    // Configured hostname
	Username           string    // Configured username
	TargetDisk         string    // Target installation disk
	EncryptionType     string    // Encryption type (none, LUKS, LUKS-LVM)
	Progress           int       // Progress percentage (0-100)
	StartedAt          *time.Time // When installation started
	CompletedAt        *time.Time // When installation completed
	CurrentPhase       string    // Current phase name
	LastError          string    // Last error message (if any)
	EstimatedRemaining int       // Estimated remaining time in seconds
}
