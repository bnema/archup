package system

import (
	"fmt"
	"strings"
)

// ChrootExec executes a command inside a chroot environment
func ChrootExec(logPath, mountPoint, command string, args ...string) error {
	// Construct the command to run in chroot
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", fullCommand)
	if result.Error != nil {
		return fmt.Errorf("failed to execute in chroot: %w", result.Error)
	}
	return nil
}

// ChrootPacman runs pacman inside the chroot
func ChrootPacman(logPath, mountPoint, operation string, packages ...string) error {
	args := append([]string{operation, "--noconfirm"}, packages...)
	command := "pacman " + strings.Join(args, " ")

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", command)
	if result.Error != nil {
		return fmt.Errorf("failed to run pacman in chroot: %w", result.Error)
	}
	return nil
}

// ChrootSystemctl runs systemctl inside the chroot
func ChrootSystemctl(logPath, mountPoint, action, service string) error {
	command := fmt.Sprintf("systemctl %s %s", action, service)

	result := RunLogged(logPath, "arch-chroot", mountPoint, "bash", "-c", command)
	if result.Error != nil {
		return fmt.Errorf("failed to run systemctl in chroot: %w", result.Error)
	}
	return nil
}

// Mount mounts a filesystem
func Mount(logPath, device, mountPoint string, options ...string) error {
	args := []string{device, mountPoint}
	if len(options) > 0 {
		args = append([]string{"-o", strings.Join(options, ",")}, args...)
	}

	result := RunLogged(logPath, "mount", args...)
	if result.Error != nil {
		return fmt.Errorf("failed to mount %s: %w", device, result.Error)
	}
	return nil
}

// Unmount unmounts a filesystem
func Unmount(logPath, mountPoint string) error {
	result := RunLogged(logPath, "umount", "-R", mountPoint)
	if result.Error != nil {
		return fmt.Errorf("failed to unmount %s: %w", mountPoint, result.Error)
	}
	return nil
}

// Genfstab generates fstab for the installed system
func Genfstab(logPath, mountPoint, fstabPath string) error {
	result := RunLogged(logPath, "genfstab", "-U", mountPoint)
	if result.Error != nil {
		return fmt.Errorf("failed to generate fstab: %w", result.Error)
	}

	// TODO: Write result.Output to fstabPath
	// This should be done using file operations
	return nil
}
