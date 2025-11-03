package disk

import (
	"errors"
	"fmt"
	"strings"
)

// Errors specific to disk validation
var (
	// ErrInvalidDiskPath is returned when disk path validation fails
	ErrInvalidDiskPath = errors.New("invalid disk path")

	// ErrInvalidPartitionPath is returned when partition path validation fails
	ErrInvalidPartitionPath = errors.New("invalid partition path")
)

// DeterminePartitionPath determines the correct partition device path based on disk type
// For NVMe disks: /dev/nvme0n1 → /dev/nvme0n1p1, /dev/nvme0n1p2
// For SATA/SSD: /dev/sda → /dev/sda1, /dev/sda2
// For MMC: /dev/mmcblk0 → /dev/mmcblk0p1, /dev/mmcblk0p2
func DeterminePartitionPath(diskPath string, partitionNumber int) (string, error) {
	if diskPath == "" {
		return "", fmt.Errorf("%w: disk path cannot be empty", ErrInvalidDiskPath)
	}

	if partitionNumber < 1 || partitionNumber > 128 {
		return "", fmt.Errorf("%w: partition number must be between 1 and 128", ErrInvalidPartitionPath)
	}

	// Check if disk path is valid
	if !strings.HasPrefix(diskPath, "/dev/") {
		return "", fmt.Errorf("%w: disk path must start with /dev/", ErrInvalidDiskPath)
	}

	// NVMe disks use 'p' prefix: nvme0n1p1, nvme0n1p2
	if strings.Contains(diskPath, "nvme") {
		return fmt.Sprintf("%sp%d", diskPath, partitionNumber), nil
	}

	// MMC disks use 'p' prefix: mmcblk0p1, mmcblk0p2
	if strings.Contains(diskPath, "mmcblk") {
		return fmt.Sprintf("%sp%d", diskPath, partitionNumber), nil
	}

	// SATA/SSD disks use direct numbering: sda1, sda2, vda1, vda2
	return fmt.Sprintf("%s%d", diskPath, partitionNumber), nil
}

// ValidateDiskPath validates a disk device path
func ValidateDiskPath(diskPath string) error {
	if diskPath == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidDiskPath)
	}

	if !strings.HasPrefix(diskPath, "/dev/") {
		return fmt.Errorf("%w: path must start with /dev/", ErrInvalidDiskPath)
	}

	// Check for common disk patterns
	validPrefixes := []string{
		"/dev/sd",     // SATA/SCSI
		"/dev/hd",     // IDE (legacy)
		"/dev/vd",     // Virtio
		"/dev/nvme",   // NVMe
		"/dev/mmcblk", // MMC
	}

	valid := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(diskPath, prefix) {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("%w: unrecognized disk device type", ErrInvalidDiskPath)
	}

	// Length check
	if len(diskPath) > 64 {
		return fmt.Errorf("%w: path too long (max 64 chars)", ErrInvalidDiskPath)
	}

	return nil
}

// ValidatePartitionPath validates a partition device path
func ValidatePartitionPath(partitionPath string) error {
	if partitionPath == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPartitionPath)
	}

	if !strings.HasPrefix(partitionPath, "/dev/") {
		return fmt.Errorf("%w: path must start with /dev/", ErrInvalidPartitionPath)
	}

	// Length check
	if len(partitionPath) > 64 {
		return fmt.Errorf("%w: path too long (max 64 chars)", ErrInvalidPartitionPath)
	}

	return nil
}

// IsDiskPath checks if the given path looks like a disk (not a partition)
func IsDiskPath(path string) bool {
	// Partitions usually end with a digit or 'p' + digit
	// Disks typically don't (except nvme and mmcblk which end with a digit)
	if strings.Contains(path, "nvme") {
		// NVMe disk: /dev/nvme0n1
		// NVMe partition: /dev/nvme0n1p1
		return !strings.Contains(path, "p")
	}

	if strings.Contains(path, "mmcblk") {
		// MMC disk: /dev/mmcblk0
		// MMC partition: /dev/mmcblk0p1
		return !strings.Contains(path, "p")
	}

	// For SATA/SCSI: /dev/sda (disk) vs /dev/sda1 (partition)
	// Check if last character is a digit
	if len(path) == 0 {
		return false
	}

	lastChar := path[len(path)-1]
	return !(lastChar >= '0' && lastChar <= '9')
}

// ExtractDiskFromPartition extracts the disk path from a partition path
// Examples:
//   - /dev/sda1 → /dev/sda
//   - /dev/nvme0n1p1 → /dev/nvme0n1
//   - /dev/mmcblk0p1 → /dev/mmcblk0
func ExtractDiskFromPartition(partitionPath string) (string, error) {
	if partitionPath == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPartitionPath)
	}

	// NVMe partitions: must have 'p' separator (e.g., nvme0n1p1)
	// If no 'p', it's already a disk (e.g., nvme0n1)
	if strings.Contains(partitionPath, "nvme") {
		if strings.Contains(partitionPath, "p") {
			// Find last 'p' and remove everything after it
			lastP := strings.LastIndex(partitionPath, "p")
			return partitionPath[:lastP], nil
		}
		// No 'p' means it's already a disk path
		return partitionPath, nil
	}

	// MMC partitions: must have 'p' separator (e.g., mmcblk0p1)
	// If no 'p', it's already a disk (e.g., mmcblk0)
	if strings.Contains(partitionPath, "mmcblk") {
		if strings.Contains(partitionPath, "p") {
			// Find last 'p' and remove everything after it
			lastP := strings.LastIndex(partitionPath, "p")
			return partitionPath[:lastP], nil
		}
		// No 'p' means it's already a disk path
		return partitionPath, nil
	}

	// SATA/SCSI: remove trailing digits
	// Find where digits start from the end
	i := len(partitionPath) - 1
	for i >= 0 && partitionPath[i] >= '0' && partitionPath[i] <= '9' {
		i--
	}

	if i == len(partitionPath)-1 {
		// No digits found - might already be a disk path
		return partitionPath, nil
	}

	return partitionPath[:i+1], nil
}
