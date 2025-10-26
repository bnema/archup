package validation

import (
	"errors"
	"fmt"
)

// ValidateNotEmpty validates that a string is not empty
func ValidateNotEmpty(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s cannot be empty", field)
		}
		return nil
	}
}

// ValidateHostname validates a hostname
func ValidateHostname(s string) error {
	if s == "" {
		return errors.New("hostname cannot be empty")
	}
	if len(s) > 63 {
		return errors.New("hostname must be 63 characters or less")
	}
	// Basic validation
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
			return errors.New("hostname can only contain letters, numbers, and hyphens")
		}
	}
	return nil
}

// ValidateUsername validates a username
func ValidateUsername(s string) error {
	if s == "" {
		return errors.New("username cannot be empty")
	}
	if len(s) > 32 {
		return errors.New("username must be 32 characters or less")
	}
	// Must start with lowercase letter
	if s[0] < 'a' || s[0] > 'z' {
		return errors.New("username must start with a lowercase letter")
	}
	// Can only contain lowercase letters, numbers, underscore, hyphen
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return errors.New("username can only contain lowercase letters, numbers, underscore, and hyphen")
		}
	}
	return nil
}

// ValidatePassword validates a password (minimum length)
func ValidatePassword(s string) error {
	if len(s) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

// ValidateTimezone validates a timezone (basic check)
func ValidateTimezone(s string) error {
	if s == "" {
		return errors.New("timezone cannot be empty")
	}
	// Basic validation - should contain /
	// More thorough validation would check against zoneinfo database
	return nil
}

// ValidateLocale validates a locale (basic check)
func ValidateLocale(s string) error {
	if s == "" {
		return errors.New("locale cannot be empty")
	}
	// Basic validation - should match pattern like en_US.UTF-8
	return nil
}
