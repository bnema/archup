package user

import "errors"

// Credentials is an immutable value object for user credentials
type Credentials struct {
	userPassword string
	rootPassword string
}

// NewCredentials creates new credentials with validation
func NewCredentials(userPassword, rootPassword string) (*Credentials, error) {
	// Validate user password
	if err := ValidatePassword(userPassword); err != nil {
		return nil, err
	}

	// Validate root password
	if err := ValidatePassword(rootPassword); err != nil {
		return nil, err
	}

	// Business rule: passwords must differ
	if userPassword == rootPassword {
		return nil, errors.New("user and root passwords must differ")
	}

	return &Credentials{
		userPassword: userPassword,
		rootPassword: rootPassword,
	}, nil
}

// HasPassword returns true if password is set
func (c *Credentials) HasPassword() bool {
	return c.userPassword != "" && c.rootPassword != ""
}

// String returns a redacted representation
func (c *Credentials) String() string {
	return "Credentials(set=true)"
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Minimum 8 characters (reduced from 12 to be more user-friendly)
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Maximum 128 characters (reasonable limit)
	if len(password) > 128 {
		return errors.New("password cannot exceed 128 characters")
	}

	return nil
}
