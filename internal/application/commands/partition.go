package commands

import "github.com/bnema/archup/internal/domain/disk"

// PartitionDiskCommand contains data for disk partitioning
type PartitionDiskCommand struct {
	TargetDisk         string              // e.g., "/dev/sda"
	RootSizeGB         int64               // Size of root partition in GB
	BootSizeGB         int64               // Size of boot partition in GB (4GB recommended for limine-snapper-sync)
	EncryptionType     disk.EncryptionType // EncryptionTypeNone, EncryptionTypeLUKS, EncryptionTypeLUKSLVM
	EncryptionPassword string              // Password for encrypted partitions (if applicable)
	FilesystemType     disk.FilesystemType // FilesystemExt4, FilesystemBtrfs, FilesystemFAT32
	WipeDisks          bool                // Whether to wipe entire disk before partitioning
}
