package disk

import (
	"errors"
	"fmt"
	"strings"
)

// MountOptions is an immutable value object representing filesystem mount options
type MountOptions struct {
	options map[string]string // key-value pairs like "compress": "zstd", "noatime": ""
}

// NewMountOptions creates a new MountOptions with validation
func NewMountOptions(options map[string]string) (*MountOptions, error) {
	if options == nil {
		return nil, errors.New("mount options cannot be nil")
	}

	// Validate each option
	for key, value := range options {
		if key == "" {
			return nil, errors.New("mount option key cannot be empty")
		}

		// Check for invalid characters in keys
		if strings.ContainsAny(key, " ,=") {
			return nil, fmt.Errorf("invalid mount option key: %s", key)
		}

		// Check for invalid characters in values
		if strings.ContainsAny(value, " ,") {
			return nil, fmt.Errorf("invalid mount option value for key %s: %s", key, value)
		}
	}

	// Create a copy to ensure immutability
	optionsCopy := make(map[string]string, len(options))
	for k, v := range options {
		optionsCopy[k] = v
	}

	return &MountOptions{
		options: optionsCopy,
	}, nil
}

// NewBtrfsMountOptions creates standard Btrfs mount options
// Used by the archup installer: noatime, compress=zstd
func NewBtrfsMountOptions(subvolume string) (*MountOptions, error) {
	options := map[string]string{
		"noatime":  "",
		"compress": "zstd",
	}

	if subvolume != "" {
		options["subvol"] = subvolume
	}

	return NewMountOptions(options)
}

// Options returns a copy of the options map
func (m *MountOptions) Options() map[string]string {
	optionsCopy := make(map[string]string, len(m.options))
	for k, v := range m.options {
		optionsCopy[k] = v
	}
	return optionsCopy
}

// Has checks if a specific option is set
func (m *MountOptions) Has(key string) bool {
	_, exists := m.options[key]
	return exists
}

// Get returns the value for a specific option
func (m *MountOptions) Get(key string) (string, bool) {
	value, exists := m.options[key]
	return value, exists
}

// ToString converts mount options to a comma-separated string
// Format: "option1,option2=value,option3=value"
func (m *MountOptions) ToString() string {
	parts := make([]string, 0, len(m.options))

	for key, value := range m.options {
		if value == "" {
			parts = append(parts, key)
		} else {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return strings.Join(parts, ",")
}

// String returns human-readable representation
func (m *MountOptions) String() string {
	return fmt.Sprintf("MountOptions(%s)", m.ToString())
}

// Equals checks if two MountOptions are equal
func (m *MountOptions) Equals(other *MountOptions) bool {
	if other == nil {
		return false
	}

	if len(m.options) != len(other.options) {
		return false
	}

	for key, value := range m.options {
		otherValue, exists := other.options[key]
		if !exists || value != otherValue {
			return false
		}
	}

	return true
}

// ParseMountOptions parses a comma-separated mount options string
// Format: "option1,option2=value,option3=value"
func ParseMountOptions(optionsStr string) (*MountOptions, error) {
	if optionsStr == "" {
		return NewMountOptions(map[string]string{})
	}

	options := make(map[string]string)
	parts := strings.Split(optionsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by = for key=value format
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid mount option format: %s", part)
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			if key == "" {
				return nil, fmt.Errorf("empty option key in: %s", part)
			}

			options[key] = value
		} else {
			// Simple flag option (no value)
			options[part] = ""
		}
	}

	return NewMountOptions(options)
}

// MountPoint is an immutable value object representing a filesystem mount point
type MountPoint struct {
	path    string
	device  string // e.g., /dev/sda1 or /dev/mapper/cryptroot
	options *MountOptions
}

// NewMountPoint creates a new MountPoint with validation
func NewMountPoint(path, device string, options *MountOptions) (*MountPoint, error) {
	// Validate path
	if path == "" {
		return nil, errors.New("mount point path cannot be empty")
	}

	if !strings.HasPrefix(path, "/") {
		return nil, errors.New("mount point path must start with /")
	}

	// Validate device
	if device == "" {
		return nil, errors.New("mount device cannot be empty")
	}

	// Validate options
	if options == nil {
		return nil, errors.New("mount options cannot be nil")
	}

	return &MountPoint{
		path:    path,
		device:  device,
		options: options,
	}, nil
}

// Path returns the mount point path
func (mp *MountPoint) Path() string {
	return mp.path
}

// Device returns the mount device
func (mp *MountPoint) Device() string {
	return mp.device
}

// Options returns the mount options
func (mp *MountPoint) Options() *MountOptions {
	return mp.options
}

// IsRoot returns true if this is the root mount point
func (mp *MountPoint) IsRoot() bool {
	return mp.path == "/"
}

// String returns human-readable representation
func (mp *MountPoint) String() string {
	return fmt.Sprintf("MountPoint(path=%s, device=%s, options=%s)",
		mp.path, mp.device, mp.options.ToString())
}

// Equals checks if two MountPoint objects are equal
func (mp *MountPoint) Equals(other *MountPoint) bool {
	if other == nil {
		return false
	}
	return mp.path == other.path &&
		mp.device == other.device &&
		mp.options.Equals(other.options)
}

// Errors specific to mount operations
var (
	// ErrInvalidMountPoint is returned when mount point validation fails
	ErrInvalidMountPoint = errors.New("invalid mount point")

	// ErrInvalidMountOptions is returned when mount options validation fails
	ErrInvalidMountOptions = errors.New("invalid mount options")
)

// ValidateMountPoint validates a mount point path
func ValidateMountPoint(path string) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidMountPoint)
	}

	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("%w: path must start with /", ErrInvalidMountPoint)
	}

	// Check for obviously invalid patterns
	if strings.Contains(path, "//") || strings.Contains(path, "..") {
		return fmt.Errorf("%w: path contains invalid patterns", ErrInvalidMountPoint)
	}

	// Length check (reasonable limit)
	if len(path) > 256 {
		return fmt.Errorf("%w: path too long (max 256 chars)", ErrInvalidMountPoint)
	}

	return nil
}
