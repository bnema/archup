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

// DownloadAndInstallPackages downloads packages from URLs using curl and installs them in chroot
func DownloadAndInstallPackages(logPath, chrootPath string, urls ...string) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	// Download and install in one command
	cmd := "mkdir -p /tmp && "
	for i, url := range urls {
		cmd += fmt.Sprintf("curl -L -f -o /tmp/pkg%d.tar.zst '%s' && ", i, url)
	}
	cmd += "pacman -U --noconfirm /tmp/pkg*.tar.zst && rm -f /tmp/pkg*.tar.zst"

	if err := ChrootExec(logPath, chrootPath, cmd); err != nil {
		ChrootExec(logPath, chrootPath, "rm -f /tmp/pkg*.tar.zst") // Cleanup
		return err
	}

	return nil
}
