package disk

import (
	"errors"
	"fmt"
)

// Disk is an entity representing a physical disk with partitions
// It manages the collection of partitions and validates partition layout
type Disk struct {
	device     string
	sizeGB     int64
	partitions []*Partition
}

// NewDisk creates a new Disk entity with validation
func NewDisk(device string, sizeGB int64) (*Disk, error) {
	// Validate device
	if device == "" {
		return nil, errors.New("disk device cannot be empty")
	}

	// Validate size
	if sizeGB < 1 {
		return nil, errors.New("disk size must be at least 1GB")
	}

	// Reasonable limit (max 100TB)
	if sizeGB > 104857600 {
		return nil, errors.New("disk size is unreasonably large")
	}

	return &Disk{
		device:     device,
		sizeGB:     sizeGB,
		partitions: []*Partition{},
	}, nil
}

// Device returns the device path
func (d *Disk) Device() string {
	return d.device
}

// SizeGB returns the disk size in gigabytes
func (d *Disk) SizeGB() int64 {
	return d.sizeGB
}

// SizeMB returns the disk size in megabytes
func (d *Disk) SizeMB() int64 {
	return d.sizeGB * 1024
}

// Partitions returns a copy of the partitions list
func (d *Disk) Partitions() []*Partition {
	partitions := make([]*Partition, len(d.partitions))
	copy(partitions, d.partitions)
	return partitions
}

// PartitionCount returns the number of partitions
func (d *Disk) PartitionCount() int {
	return len(d.partitions)
}

// AddPartition adds a partition to the disk with validation
func (d *Disk) AddPartition(partition *Partition) error {
	if partition == nil {
		return errors.New("partition cannot be nil")
	}

	// Check partition device is on this disk
	if !isPartitionOfDisk(d.device, partition.device) {
		return fmt.Errorf("partition %s is not on disk %s", partition.device, d.device)
	}

	// Check for duplicate device
	for _, p := range d.partitions {
		if p.device == partition.device {
			return fmt.Errorf("partition %s already exists", partition.device)
		}
	}

	// Check for duplicate mount points (except swap which has no mount point)
	if partition.mountPoint != "" {
		for _, p := range d.partitions {
			if p.mountPoint == partition.mountPoint {
				return fmt.Errorf("mount point %s is already in use", partition.mountPoint)
			}
		}
	}

	// Validate total size doesn't exceed disk size
	if totalPartitionSize := d.calculateTotalPartitionSize() + partition.SizeGB(); totalPartitionSize > d.sizeGB {
		return fmt.Errorf("partition size would exceed disk capacity (total: %dGB, disk: %dGB)",
			totalPartitionSize, d.sizeGB)
	}

	d.partitions = append(d.partitions, partition)
	return nil
}

// FindPartitionByMountPoint finds a partition by its mount point
func (d *Disk) FindPartitionByMountPoint(mountPoint string) *Partition {
	for _, p := range d.partitions {
		if p.mountPoint == mountPoint {
			return p
		}
	}
	return nil
}

// FindPartitionByDevice finds a partition by its device path
func (d *Disk) FindPartitionByDevice(device string) *Partition {
	for _, p := range d.partitions {
		if p.device == device {
			return p
		}
	}
	return nil
}

// HasRootPartition returns true if disk has a root (/) partition
func (d *Disk) HasRootPartition() bool {
	return d.FindPartitionByMountPoint("/") != nil
}

// HasBootPartition returns true if disk has a boot partition
func (d *Disk) HasBootPartition() bool {
	return d.FindPartitionByMountPoint("/boot") != nil || d.FindPartitionByMountPoint("/boot/efi") != nil
}

// HasEFIPartition returns true if disk has an EFI partition
func (d *Disk) HasEFIPartition() bool {
	return d.FindPartitionByMountPoint("/boot/efi") != nil
}

// GetEFIPartition returns the EFI partition if it exists
func (d *Disk) GetEFIPartition() *Partition {
	return d.FindPartitionByMountPoint("/boot/efi")
}

// GetRootPartition returns the root partition if it exists
func (d *Disk) GetRootPartition() *Partition {
	return d.FindPartitionByMountPoint("/")
}

// GetBootPartition returns the boot partition if it exists
func (d *Disk) GetBootPartition() *Partition {
	if p := d.FindPartitionByMountPoint("/boot/efi"); p != nil {
		return p
	}
	return d.FindPartitionByMountPoint("/boot")
}

// CalculateUsedSpace returns total space used by partitions in GB
func (d *Disk) CalculateUsedSpace() int64 {
	return d.calculateTotalPartitionSize()
}

// CalculateFreeSpace returns remaining space in GB
func (d *Disk) CalculateFreeSpace() int64 {
	return d.sizeGB - d.calculateTotalPartitionSize()
}

// ValidateLayout validates the partition layout according to business rules
func (d *Disk) ValidateLayout() error {
	// At minimum, must have root partition
	if !d.HasRootPartition() {
		return errors.New("disk must have a root (/) partition")
	}

	rootPartition := d.GetRootPartition()

	// Root partition must be at least 20GB for comfortable installation
	if rootPartition.SizeGB() < 20 {
		return errors.New("root partition must be at least 20GB")
	}

	// For UEFI boot, must have EFI partition
	// (This is a business rule - we assume UEFI for now)
	if !d.HasEFIPartition() {
		return errors.New("disk must have an EFI partition (/boot/efi) for UEFI boot")
	}

	// EFI partition must be FAT32
	efiPartition := d.GetEFIPartition()
	if efiPartition.Filesystem() != FilesystemFAT32 {
		return errors.New("EFI partition must use FAT32 filesystem")
	}

	// EFI partition must be at least 256MB (typical 512MB)
	if efiPartition.SizeGB() < 1 {
		// Less than 1GB is ok for EFI if it's at least 256MB
		if efiPartition.SizeMB() < 256 {
			return errors.New("EFI partition must be at least 256MB")
		}
	}

	// No duplicate mount points
	mountPoints := make(map[string]bool)
	for _, p := range d.partitions {
		if p.mountPoint != "" {
			if mountPoints[p.mountPoint] {
				return fmt.Errorf("duplicate mount point: %s", p.mountPoint)
			}
			mountPoints[p.mountPoint] = true
		}
	}

	return nil
}

// String returns human-readable representation
func (d *Disk) String() string {
	return fmt.Sprintf("Disk(device=%s, size=%dGB, partitions=%d)", d.device, d.sizeGB, len(d.partitions))
}

// Private helper methods

func (d *Disk) calculateTotalPartitionSize() int64 {
	var total int64
	for _, p := range d.partitions {
		total += p.SizeGB()
	}
	return total
}

// isPartitionOfDisk checks if a partition device belongs to a disk
// e.g., /dev/sda1 belongs to /dev/sda, /dev/nvme0n1p1 belongs to /dev/nvme0n1
func isPartitionOfDisk(diskDevice, partitionDevice string) bool {
	if len(partitionDevice) <= len(diskDevice) {
		return false
	}

	// Check prefix matches
	if partitionDevice[:len(diskDevice)] != diskDevice {
		return false
	}

	// Next character can be 'p' (for nvme: nvme0n1p1) or a digit (for sda: sda1)
	nextChar := partitionDevice[len(diskDevice)]
	return (nextChar >= '0' && nextChar <= '9') || nextChar == 'p'
}
