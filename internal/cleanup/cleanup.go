package cleanup

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
)

// Run performs a comprehensive cleanup of installation artifacts.
// This function is idempotent and safe to call multiple times.
// It performs the following operations:
// 1. Unmount /mnt hierarchy (recursive, force if needed)
// 2. Close LUKS/dm-crypt mappings (cryptroot and all others)
// 3. Disable swap
// 4. Kill stuck pacstrap/arch-chroot processes
// 5. Remove installation config file
// 6. Sync filesystems
func Run(log *logger.Logger) error {
	log.Info("Starting comprehensive cleanup")

	cleanupMounts(log)
	cleanupEncryption(log)
	cleanupSwap(log)
	cleanupProcesses(log)
	cleanupConfig(log)
	cleanupSync(log)

	log.Info("Cleanup complete")
	return nil
}

// cleanupMounts unmounts the /mnt hierarchy in reverse order.
func cleanupMounts(log *logger.Logger) {
	log.Info("Cleaning up mounted filesystems")

	// Unmount in reverse order (most specific first)
	mountsToClean := []string{
		"/mnt/boot",
		"/mnt/home",
		"/mnt",
	}

	for _, mountPoint := range mountsToClean {
		// Check if mount point is mounted
		if !isMounted(mountPoint) {
			continue
		}

		log.Info("Unmounting", "mount_point", mountPoint)

		// Try normal unmount first
		cmd := exec.Command("umount", "-R", mountPoint)
		if err := cmd.Run(); err == nil {
			log.Info("Unmounted successfully", "mount_point", mountPoint)
			continue
		}

		// Force unmount if normal unmount fails
		log.Info("Force unmounting", "mount_point", mountPoint)
		cmd = exec.Command("umount", "-R", "-l", mountPoint)
		if err := cmd.Run(); err == nil {
			log.Info("Force unmounted successfully", "mount_point", mountPoint)
		} else {
			log.Warn("Failed to unmount (may retry later)", "mount_point", mountPoint, "error", err)
		}
	}

	// Wait for any lingering unmount operations
	time.Sleep(1 * time.Second)
}

// cleanupEncryption closes LUKS/dm-crypt mappings.
func cleanupEncryption(log *logger.Logger) {
	log.Info("Cleaning up LUKS/dm-crypt mappings")

	// Close cryptroot specifically
	if _, err := os.Stat("/dev/mapper/cryptroot"); err == nil {
		log.Info("Closing LUKS container", "device", "cryptroot")
		cmd := exec.Command("cryptsetup", "close", "cryptroot")
		if err := cmd.Run(); err == nil {
			log.Info("Closed LUKS container successfully", "device", "cryptroot")
		} else {
			log.Warn("Failed to close LUKS container", "device", "cryptroot", "error", err)
		}
	}

	// Close ALL dm-crypt mappings
	cmd := exec.Command("dmsetup", "ls", "--target", "crypt")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			// Extract mapping name (first field)
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			mapping := fields[0]

			log.Info("Closing dm-crypt mapping", "device", mapping)
			closeCmd := exec.Command("cryptsetup", "close", mapping)
			if err := closeCmd.Run(); err == nil {
				log.Info("Closed dm-crypt mapping successfully", "device", mapping)
			} else {
				log.Warn("Failed to close dm-crypt mapping", "device", mapping, "error", err)
			}
		}
	}

	// Ensure all device mapper entries are removed
	cmd = exec.Command("dmsetup", "table")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), "crypt") {
		log.Info("Removing lingering device mapper entries")
		removeCmd := exec.Command("dmsetup", "remove_all")
		_ = removeCmd.Run() // Ignore errors
		log.Info("Device mapper cleanup attempted")
	}
}

// cleanupSwap disables all active swap.
func cleanupSwap(log *logger.Logger) {
	log.Info("Disabling swap")

	// Check if swap is active
	cmd := exec.Command("swapon", "--show")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		log.Info("Found active swap, disabling")
		swapoffCmd := exec.Command("swapoff", "-a")
		if err := swapoffCmd.Run(); err == nil {
			log.Info("Swap disabled successfully")
		} else {
			log.Warn("Failed to disable swap (non-critical)", "error", err)
		}
	}
}

// cleanupProcesses kills stuck pacstrap/arch-chroot processes.
func cleanupProcesses(log *logger.Logger) {
	log.Info("Checking for stuck pacstrap/chroot processes")

	// Find stuck processes
	cmd := exec.Command("pgrep", "-f", "pacstrap|arch-chroot")
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns 1 if no processes found, which is expected/normal
		return
	}

	if len(output) > 0 {
		pids := strings.TrimSpace(string(output))
		log.Info("Killing stuck processes", "pids", pids)

		// Try SIGTERM first
		pidList := strings.Fields(pids)
		for _, pid := range pidList {
			killCmd := exec.Command("kill", pid)
			_ = killCmd.Run() // Ignore errors
		}

		time.Sleep(1 * time.Second)

		// Try SIGKILL if still running
		for _, pid := range pidList {
			killCmd := exec.Command("kill", "-9", pid)
			_ = killCmd.Run() // Ignore errors
		}

		log.Info("Cleaned up stuck processes")
	}
}

// cleanupConfig removes the installation config file.
func cleanupConfig(log *logger.Logger) {
	log.Info("Removing installation config file")

	configPath := config.DefaultConfigPath
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err == nil {
			log.Info("Config file removed", "path", configPath)
		} else {
			log.Warn("Failed to remove config file", "path", configPath, "error", err)
		}
	}
}

// cleanupSync syncs all filesystems.
func cleanupSync(log *logger.Logger) {
	log.Info("Syncing filesystems")

	cmd := exec.Command("sync")
	if err := cmd.Run(); err == nil {
		log.Info("Filesystems synced successfully")
	} else {
		log.Warn("Failed to sync filesystems (non-critical)", "error", err)
	}
}

// isMounted checks if a mount point is currently mounted.
func isMounted(mountPoint string) bool {
	cmd := exec.Command("mountpoint", "-q", mountPoint)
	return cmd.Run() == nil
}
