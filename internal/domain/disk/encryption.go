package disk

import (
	"errors"

	"github.com/bnema/archup/internal/domain/user"
)

// EncryptionType represents the disk encryption method
type EncryptionType int

const (
	// EncryptionTypeNone means no encryption
	EncryptionTypeNone EncryptionType = iota

	// EncryptionTypeLUKS means LUKS encryption (basic)
	EncryptionTypeLUKS

	// EncryptionTypeLUKSLVM means LUKS with LVM
	EncryptionTypeLUKSLVM
)

// String returns human-readable encryption type name
func (e EncryptionType) String() string {
	switch e {
	case EncryptionTypeLUKS:
		return "LUKS"
	case EncryptionTypeLUKSLVM:
		return "LUKS+LVM"
	default:
		return "None"
	}
}

// IsEncrypted returns true if encryption is enabled
func (e EncryptionType) IsEncrypted() bool {
	return e != EncryptionTypeNone
}

// EncryptionConfig is an immutable value object representing encryption configuration
type EncryptionConfig struct {
	encType  EncryptionType
	password string
}

// NewEncryptionConfig creates a new encryption configuration with validation
func NewEncryptionConfig(encType EncryptionType, password string) (*EncryptionConfig, error) {
	// If encryption is enabled, password is required
	if encType.IsEncrypted() && password == "" {
		return nil, errors.New("encryption password required when encryption is enabled")
	}

	// If encryption is disabled, password should be empty
	if !encType.IsEncrypted() && password != "" {
		return nil, errors.New("password cannot be set when encryption is disabled")
	}

	// Validate password strength if encryption is enabled
	if encType.IsEncrypted() {
		if err := validateEncryptionPassword(password); err != nil {
			return nil, err
		}
	}

	return &EncryptionConfig{
		encType:  encType,
		password: password,
	}, nil
}

// EncryptionType returns the encryption type
func (e *EncryptionConfig) EncryptionType() EncryptionType {
	return e.encType
}

// IsEncrypted returns true if encryption is enabled
func (e *EncryptionConfig) IsEncrypted() bool {
	return e.encType.IsEncrypted()
}

// HasPassword returns true if a password is set
// (always true if encrypted, always false if not)
func (e *EncryptionConfig) HasPassword() bool {
	return e.password != ""
}

// String returns human-readable representation
func (e *EncryptionConfig) String() string {
	return "EncryptionConfig(type=" + e.encType.String() + ", encrypted=" + boolToString(e.IsEncrypted()) + ")"
}

// Equals checks if two EncryptionConfig objects are equal
// Note: Password comparison is constant-time safe
func (e *EncryptionConfig) Equals(other *EncryptionConfig) bool {
	if other == nil {
		return false
	}
	return e.encType == other.encType && constantTimeEqual(e.password, other.password)
}

// Private helper functions

func validateEncryptionPassword(password string) error {
	// Delegate to user password validation (single source of truth)
	return user.ValidatePassword(password)
}

// constantTimeEqual compares two strings in constant time to prevent timing attacks
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
