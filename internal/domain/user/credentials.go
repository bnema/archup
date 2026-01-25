package user

import "errors"

// Credentials is an immutable value object for user credentials
type Credentials struct {
	userPassword string
	rootPassword string
}

// NewCredentials creates new credentials with validation
// rootPassword can be empty (root account will be locked, user uses sudo)
func NewCredentials(userPassword, rootPassword string) (*Credentials, error) {
	// Validate user password (required)
	if err := ValidatePassword(userPassword); err != nil {
		return nil, err
	}

	// Root password is optional - empty means root account locked
	if rootPassword != "" {
		if err := ValidatePassword(rootPassword); err != nil {
			return nil, err
		}
		// Business rule: if root password set, it must differ from user password
		if userPassword == rootPassword {
			return nil, errors.New("user and root passwords must differ")
		}
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

	// Minimum 4 characters
	if len(password) < 4 {
		return errors.New("password must be at least 4 characters")
	}

	// Maximum 128 characters (reasonable limit)
	if len(password) > 128 {
		return errors.New("password cannot exceed 128 characters")
	}

	return nil
}
