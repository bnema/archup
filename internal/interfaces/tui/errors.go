package tui

import "errors"

// TUI validation and runtime errors
var (
	ErrInvalidHostname = errors.New("hostname is required and must be valid")
	ErrInvalidUsername = errors.New("username is required")
	ErrInvalidDisk     = errors.New("target disk is required")
	ErrInvalidPassword = errors.New("password must be at least 3 characters")
	ErrDiskNotFound    = errors.New("target disk not found")
	ErrFormIncomplete  = errors.New("form is incomplete, please fill all required fields")
)
