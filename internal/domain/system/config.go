package system

import (
	"errors"
	"strings"
)

// SystemConfig is an immutable value object representing system configuration
// It encapsulates hostname, locale, timezone, and keymap settings
type SystemConfig struct {
	hostname string
	timezone string
	locale   string
	keymap   string
}

// NewSystemConfig creates a new SystemConfig with validation
func NewSystemConfig(hostname, timezone, locale, keymap string) (*SystemConfig, error) {
	// Validate hostname
	if err := ValidateHostname(hostname); err != nil {
		return nil, err
	}

	// Validate timezone
	if err := ValidateTimezone(timezone); err != nil {
		return nil, err
	}

	// Validate locale
	if err := ValidateLocale(locale); err != nil {
		return nil, err
	}

	// Validate keymap
	if err := ValidateKeymap(keymap); err != nil {
		return nil, err
	}

	return &SystemConfig{
		hostname: hostname,
		timezone: timezone,
		locale:   locale,
		keymap:   keymap,
	}, nil
}

// Hostname returns the configured hostname
func (c *SystemConfig) Hostname() string {
	return c.hostname
}

// Timezone returns the configured timezone
func (c *SystemConfig) Timezone() string {
	return c.timezone
}

// Locale returns the configured locale
func (c *SystemConfig) Locale() string {
	return c.locale
}

// Keymap returns the configured keymap
func (c *SystemConfig) Keymap() string {
	return c.keymap
}

// String returns a human-readable representation
func (c *SystemConfig) String() string {
	return "SystemConfig(" +
		"hostname=" + c.hostname +
		", timezone=" + c.timezone +
		", locale=" + c.locale +
		", keymap=" + c.keymap +
		")"
}

// Equals checks if two SystemConfig objects are equal
// Value objects should be compared by their fields
func (c *SystemConfig) Equals(other *SystemConfig) bool {
	if other == nil {
		return false
	}
	return c.hostname == other.hostname &&
		c.timezone == other.timezone &&
		c.locale == other.locale &&
		c.keymap == other.keymap
}

// Errors specific to SystemConfig
var (
	// ErrInvalidHostname is returned when hostname validation fails
	ErrInvalidHostname = errors.New("invalid hostname")

	// ErrInvalidTimezone is returned when timezone validation fails
	ErrInvalidTimezone = errors.New("invalid timezone")

	// ErrInvalidLocale is returned when locale validation fails
	ErrInvalidLocale = errors.New("invalid locale")

	// ErrInvalidKeymap is returned when keymap validation fails
	ErrInvalidKeymap = errors.New("invalid keymap")
)

// ValidateHostname validates hostname according to RFC 1123
// Valid hostnames: letters, digits, hyphens
// Cannot start/end with hyphen, length 1-63 characters
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return ErrInvalidHostname
	}

	if len(hostname) > 63 {
		return ErrInvalidHostname
	}

	if hostname[0] == '-' || hostname[len(hostname)-1] == '-' {
		return ErrInvalidHostname
	}

	for _, ch := range hostname {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-') {
			return ErrInvalidHostname
		}
	}

	return nil
}

// ValidateTimezone validates timezone string
// Common zones: "UTC", "US/Eastern", "Europe/London", etc.
// Empty timezone is allowed (means system default)
func ValidateTimezone(timezone string) error {
	// Empty timezone is allowed (system default)
	if timezone == "" {
		return nil
	}

	// Check for obviously invalid patterns
	if strings.Contains(timezone, "..") || strings.HasPrefix(timezone, "/") {
		return ErrInvalidTimezone
	}

	// Length check (reasonable limit)
	if len(timezone) > 64 {
		return ErrInvalidTimezone
	}

	return nil
}

// ValidateLocale validates locale string
// Valid formats: "en_US.UTF-8", "en_US", "C.UTF-8", etc.
// Empty locale is allowed (means system default)
func ValidateLocale(locale string) error {
	// Empty locale is allowed (system default)
	if locale == "" {
		return nil
	}

	// Basic format validation
	if len(locale) > 64 {
		return ErrInvalidLocale
	}

	// Should contain language code
	if !strings.Contains(locale, "_") && !strings.Contains(locale, ".") && locale != "C" {
		return ErrInvalidLocale
	}

	return nil
}

// ValidateKeymap validates keyboard layout
// Common layouts: us, de, fr, dvorak, etc.
// Can include variations like de-nodeadkeys
func ValidateKeymap(keymap string) error {
	if keymap == "" {
		return ErrInvalidKeymap
	}

	if len(keymap) > 64 {
		return ErrInvalidKeymap
	}

	// Allow letters, digits, hyphens
	for _, ch := range keymap {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-') {
			return ErrInvalidKeymap
		}
	}

	return nil
}
