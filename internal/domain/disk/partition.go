package disk

import (
	"errors"
	"fmt"
)

// FilesystemType represents partition filesystem type
type FilesystemType int

const (
	// FilesystemExt4 is the Linux ext4 filesystem
	FilesystemExt4 FilesystemType = iota

	// FilesystemFAT32 is the FAT32 filesystem (typically for EFI)
	FilesystemFAT32

	// FilesystemBtrfs is the Btrfs filesystem
	FilesystemBtrfs
)

// String returns human-readable filesystem type name
func (f FilesystemType) String() string {
	switch f {
	case FilesystemFAT32:
		return "FAT32"
	case FilesystemBtrfs:
		return "Btrfs"
	default:
		return "ext4"
	}
}

// Partition is an immutable value object representing a disk partition
type Partition struct {
	device     string // e.g., /dev/sda1
	size       int64  // Size in megabytes
	filesystem FilesystemType
	mountPoint string // e.g., /, /boot, /home
	encrypted  bool
}

// NewPartition creates a new Partition value object with validation
func NewPartition(device string, sizeMB int64, fs FilesystemType, mountPoint string, encrypted bool) (*Partition, error) {
	// Validate device
	if device == "" {
		return nil, errors.New("partition device cannot be empty")
	}

	// Validate size
	// Size must be positive and reasonable (less than 100TB)
	if sizeMB < 1 {
		return nil, errors.New("partition size must be at least 1MB")
	}
	if sizeMB > 104857600 { // 100TB in MB
		return nil, errors.New("partition size is unreasonably large")
	}

	// Validate mount point
	if mountPoint != "" && (mountPoint[0:1] != "/") {
		return nil, errors.New("mount point must start with / or be empty")
	}

	// Filesystem-specific validations
	if err := validateFilesystemMountPoint(fs, mountPoint); err != nil {
		return nil, err
	}

	return &Partition{
		device:     device,
		size:       sizeMB,
		filesystem: fs,
		mountPoint: mountPoint,
		encrypted:  encrypted,
	}, nil
}

// Device returns the device path
func (p *Partition) Device() string {
	return p.device
}

// SizeMB returns the partition size in megabytes
func (p *Partition) SizeMB() int64 {
	return p.size
}

// SizeGB returns the partition size in gigabytes (rounded)
func (p *Partition) SizeGB() int64 {
	return (p.size + 512) / 1024 // Round to nearest GB
}

// Filesystem returns the filesystem type
func (p *Partition) Filesystem() FilesystemType {
	return p.filesystem
}

// MountPoint returns the mount point
func (p *Partition) MountPoint() string {
	return p.mountPoint
}

// IsEncrypted returns true if partition is encrypted
func (p *Partition) IsEncrypted() bool {
	return p.encrypted
}

// IsBootPartition returns true if this is a boot partition
func (p *Partition) IsBootPartition() bool {
	return p.mountPoint == "/boot" || p.mountPoint == "/boot/efi"
}

// IsRootPartition returns true if this is the root partition
func (p *Partition) IsRootPartition() bool {
	return p.mountPoint == "/"
}

// IsEFI returns true if this is an EFI partition
func (p *Partition) IsEFI() bool {
	return p.filesystem == FilesystemFAT32 && p.mountPoint == "/boot/efi"
}

// String returns human-readable representation
func (p *Partition) String() string {
	return fmt.Sprintf("Partition(device=%s, size=%dGB, fs=%s, mount=%s, encrypted=%v)",
		p.device, p.SizeGB(), p.filesystem.String(), p.mountPoint, p.encrypted)
}

// Equals checks if two partitions are equal
func (p *Partition) Equals(other *Partition) bool {
	if other == nil {
		return false
	}
	return p.device == other.device &&
		p.size == other.size &&
		p.filesystem == other.filesystem &&
		p.mountPoint == other.mountPoint &&
		p.encrypted == other.encrypted
}

// Private helper functions

func validateFilesystemMountPoint(fs FilesystemType, mountPoint string) error {
	// FAT32 should only be used for EFI partitions
	if fs == FilesystemFAT32 && mountPoint != "" && mountPoint != "/boot/efi" {
		return errors.New("FAT32 filesystem should only be used for /boot/efi")
	}

	return nil
}

func sizeToGB(bytes int64) int64 {
	return bytes / (1024 * 1024 * 1024)
}

func sizeBytes(gb int64) int64 {
	return gb * 1024 * 1024 * 1024
}
