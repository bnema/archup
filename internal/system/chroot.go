package system

import (
	"fmt"
	"os"
	"os/exec"
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

// ChrootExecWithStdin executes a command inside chroot with stdin input
// This is secure for passing sensitive data like passwords without exposing them in process listings
func ChrootExecWithStdin(logPath, mountPoint, command, stdin string) error {
	cmd := exec.Command("arch-chroot", mountPoint, "bash", "-c", command)

	// Set up stdin pipe
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Set up logging
	if logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer logFile.Close()
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Write to stdin and close pipe
	if _, err := stdinPipe.Write([]byte(stdin)); err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdinPipe.Close()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute in chroot: %w", err)
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
