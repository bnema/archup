package user

import (
	"errors"
	"strings"
)

// User is an entity representing a system user account
type User struct {
	username string
	groups   []string
	shell    string
	home     string
}

// NewUser creates a new User entity with validation
func NewUser(username string, shell string) (*User, error) {
	// Validate username
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	// Validate shell
	if err := ValidateShell(shell); err != nil {
		return nil, err
	}

	user := &User{
		username: username,
		shell:    shell,
		groups:   []string{},
		home:     "/home/" + username,
	}

	return user, nil
}

// Username returns the username
func (u *User) Username() string {
	return u.username
}

// Shell returns the login shell
func (u *User) Shell() string {
	return u.shell
}

// Home returns the home directory path
func (u *User) Home() string {
	return u.home
}

// Groups returns a copy of the user groups list
func (u *User) Groups() []string {
	groups := make([]string, len(u.groups))
	copy(groups, u.groups)
	return groups
}

// AddGroup adds a user to a group
func (u *User) AddGroup(group string) error {
	if group == "" {
		return errors.New("group name cannot be empty")
	}

	// Check if already a member
	for _, g := range u.groups {
		if g == group {
			return errors.New("user already in group " + group)
		}
	}

	u.groups = append(u.groups, group)
	return nil
}

// HasGroup checks if user is in a group
func (u *User) HasGroup(group string) bool {
	for _, g := range u.groups {
		if g == group {
			return true
		}
	}
	return false
}

// String returns human-readable representation
func (u *User) String() string {
	return "User(username=" + u.username + ", shell=" + u.shell + ", groups=" + strings.Join(u.groups, ",") + ")"
}

// ValidateUsername validates username format
// Usernames: lowercase letters, digits, underscore, hyphen (3-32 chars)
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}

	if len(username) > 32 {
		return errors.New("username cannot exceed 32 characters")
	}

	// Must start with lowercase letter
	if username[0] < 'a' || username[0] > 'z' {
		return errors.New("username must start with lowercase letter")
	}

	// Can contain lowercase, digits, underscore, hyphen
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '_' || ch == '-') {
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}

// ValidateShell validates shell path
func ValidateShell(shell string) error {
	if shell == "" {
		return errors.New("shell cannot be empty")
	}

	if !strings.HasPrefix(shell, "/") {
		return errors.New("shell path must be absolute")
	}

	// Common valid shells
	validShells := map[string]bool{
		"/bin/bash":     true,
		"/bin/sh":       true,
		"/bin/zsh":      true,
		"/bin/fish":     true,
		"/usr/bin/zsh":  true,
		"/usr/bin/fish": true,
	}

	if validShells[shell] {
		return nil
	}

	// Allow other shells if they look reasonable
	if len(shell) > 256 {
		return errors.New("shell path too long")
	}

	return nil
}
