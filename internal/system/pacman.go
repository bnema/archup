package system

import (
	"fmt"
)

// PacmanSync syncs the package database
func PacmanSync(logPath string) error {
	result := RunLogged(logPath, "pacman", "-Sy", "--noconfirm")
	if result.Error != nil {
		return fmt.Errorf("failed to sync pacman database: %w", result.Error)
	}
	return nil
}

// PacmanInstall installs packages
func PacmanInstall(logPath string, packages ...string) error {
	args := append([]string{"-S", "--needed", "--noconfirm"}, packages...)
	result := RunLogged(logPath, "pacman", args...)
	if result.Error != nil {
		return fmt.Errorf("failed to install packages: %w", result.Error)
	}
	return nil
}

// Pacstrap installs the base system to a mount point
func Pacstrap(logPath, mountPoint string, packages ...string) error {
	args := append([]string{mountPoint}, packages...)
	result := RunLogged(logPath, "pacstrap", args...)
	if result.Error != nil {
		return fmt.Errorf("failed to run pacstrap: %w", result.Error)
	}
	return nil
}

// PacmanRemove removes packages
func PacmanRemove(logPath string, packages ...string) error {
	args := append([]string{"-R", "--noconfirm"}, packages...)
	result := RunLogged(logPath, "pacman", args...)
	if result.Error != nil {
		return fmt.Errorf("failed to remove packages: %w", result.Error)
	}
	return nil
}
