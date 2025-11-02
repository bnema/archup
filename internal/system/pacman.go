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
// Supports mirror fallbacks: URLs with the same package name but different mirrors
func DownloadAndInstallPackages(logPath, chrootPath string, urls ...string) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	// Group URLs by package (same filename = same package, different mirrors)
	// This allows fallback to alternative mirrors if primary fails
	packageMirrors := make(map[string][]string)
	for _, url := range urls {
		// Extract package name from URL
		var pkgName string
		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '/' {
				pkgName = url[i+1:]
				break
			}
		}
		packageMirrors[pkgName] = append(packageMirrors[pkgName], url)
	}

	// Download with fallback: try each mirror for each package
	cmd := "mkdir -p /tmp && "
	for i, pkgName := range getSortedKeys(packageMirrors) {
		mirrors := packageMirrors[pkgName]
		// Try primary mirror first, fallback to others if it fails
		cmd += fmt.Sprintf("(curl -L -f -o /tmp/pkg%d.tar.zst '%s'", i, mirrors[0])
		for _, mirror := range mirrors[1:] {
			cmd += fmt.Sprintf(" || curl -L -f -o /tmp/pkg%d.tar.zst '%s'", i, mirror)
		}
		cmd += ") && "
	}
	cmd += "pacman -U --noconfirm /tmp/pkg*.tar.zst && rm -f /tmp/pkg*.tar.zst"

	if err := ChrootExec(logPath, chrootPath, cmd); err != nil {
		ChrootExec(logPath, chrootPath, "rm -f /tmp/pkg*.tar.zst") // Cleanup
		return err
	}

	return nil
}

// getSortedKeys returns map keys in a stable order (important for reproducibility)
func getSortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple stable sort (not alphabetical, but consistent)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
