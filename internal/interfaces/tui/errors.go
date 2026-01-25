package tui

import "errors"

// TUI-specific errors (validation errors are in domain/system package)
var (
	ErrInvalidDisk    = errors.New("target disk is required")
	ErrDiskNotFound   = errors.New("target disk not found")
	ErrFormIncomplete = errors.New("form is incomplete, please fill all required fields")
)
