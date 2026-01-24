package dto

import "time"

// ProgressUpdate represents a progress update during installation
type ProgressUpdate struct {
	Phase           string    // Current phase name
	PhaseNumber     int       // Current phase number (1-8)
	TotalPhases     int       // Total phases (8)
	ProgressPercent int       // Overall progress percentage (0-100)
	Message         string    // Status message
	IsError         bool      // Whether this is an error
	Timestamp       time.Time // When this update occurred
}
