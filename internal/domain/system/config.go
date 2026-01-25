package system

import (
	"context"
	"errors"
	"strings"

	"github.com/bnema/archup/internal/domain/ports"
)

// DistributionType represents the type of Arch-based distribution
type DistributionType int

const (
	// DistributionArch represents vanilla Arch Linux
	DistributionArch DistributionType = iota
	// DistributionCachyOS represents CachyOS
	DistributionCachyOS
	// DistributionEndeavourOS represents EndeavourOS
	DistributionEndeavourOS
	// DistributionGaruda represents Garuda Linux
	DistributionGaruda
	// DistributionManjaro represents Manjaro Linux
	DistributionManjaro
	// DistributionUnknown represents an unknown or unsupported distribution
	DistributionUnknown
)

// String returns the string representation of the distribution type
func (d DistributionType) String() string {
	switch d {
	case DistributionArch:
		return "Arch Linux"
	case DistributionCachyOS:
		return "CachyOS"
	case DistributionEndeavourOS:
		return "EndeavourOS"
	case DistributionGaruda:
		return "Garuda Linux"
	case DistributionManjaro:
		return "Manjaro Linux"
	case DistributionUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// DetectDistribution detects which Arch-based distribution is running
func DetectDistribution(ctx context.Context, fs ports.FileSystem) (DistributionType, error) {
	// Check for vanilla Arch first
	exists, err := fs.Exists("/etc/arch-release")
	if err != nil {
		return DistributionUnknown, errors.New("failed to detect distribution: " + err.Error())
	}
	if !exists {
		return DistributionUnknown, errors.New("not running on Arch Linux or Arch ISO")
	}

	// Check for derivatives
	if exists, _ := fs.Exists("/etc/cachyos-release"); exists {
		return DistributionCachyOS, nil
	}
	if exists, _ := fs.Exists("/etc/eos-release"); exists {
		return DistributionEndeavourOS, nil
	}
	if exists, _ := fs.Exists("/etc/garuda-release"); exists {
		return DistributionGaruda, nil
	}
	if exists, _ := fs.Exists("/etc/manjaro-release"); exists {
		return DistributionManjaro, nil
	}

	// If no derivative markers found, it's vanilla Arch
	return DistributionArch, nil
}

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
	ErrInvalidHostname = errors.New("invalid hostname: must be 1-63 chars, alphanumeric and hyphens, no leading/trailing hyphen")

	// ErrInvalidTimezone is returned when timezone validation fails
	ErrInvalidTimezone = errors.New("invalid timezone")

	// ErrInvalidLocale is returned when locale validation fails
	ErrInvalidLocale = errors.New("invalid locale")

	// ErrInvalidKeymap is returned when keymap validation fails
	ErrInvalidKeymap = errors.New("invalid keymap")

	// ErrInvalidUsername is returned when username validation fails
	ErrInvalidUsername = errors.New("invalid username: must be 1-32 chars, lowercase, start with letter")

	// ErrInvalidPassword is returned when password validation fails
	ErrInvalidPassword = errors.New("invalid password: must be at least 4 characters")
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
		if !isHostnameChar(ch) {
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
		if !isKeymapChar(ch) {
			return ErrInvalidKeymap
		}
	}

	return nil
}

func isHostnameChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-'
}

func isKeymapChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-'
}

// ValidateUsername validates Linux username
// Must be 1-32 chars, lowercase, start with letter, alphanumeric and underscores
func ValidateUsername(username string) error {
	if username == "" || len(username) > 32 {
		return ErrInvalidUsername
	}

	// Must start with lowercase letter
	if username[0] < 'a' || username[0] > 'z' {
		return ErrInvalidUsername
	}

	// Must be lowercase alphanumeric or underscore
	for _, ch := range username {
		if !isUsernameChar(ch) {
			return ErrInvalidUsername
		}
	}

	return nil
}

func isUsernameChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

// ValidatePassword validates password using user domain validation
// Must be at least 4 characters
func ValidatePassword(password string) error {
	if password == "" {
		return ErrInvalidPassword
	}
	if len(password) < 4 {
		return ErrInvalidPassword
	}
	if len(password) > 128 {
		return ErrInvalidPassword
	}
	return nil
}
