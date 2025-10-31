package system

import (
	"log/slog"
)

// System provides system-level operations with logging support
type System struct {
	log *slog.Logger
}

// NewSystem creates a new System instance with the provided logger
func NewSystem(log *slog.Logger) *System {
	return &System{
		log: log,
	}
}
