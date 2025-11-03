package dto

// PartitionInfo contains information about a partition
type PartitionInfo struct {
	Device     string
	SizeGB     int64
	Filesystem string
	MountPoint string
	Encrypted  bool
}

// PartitionResult is the result of disk partitioning
type PartitionResult struct {
	TargetDisk    string
	Success       bool
	Partitions    []*PartitionInfo
	EFIPartition  string   // EFI partition path (e.g., /dev/sda1 or /dev/nvme0n1p1)
	RootPartition string   // Root partition path (e.g., /dev/sda2 or /dev/nvme0n1p2)
	CryptDevice   string   // LUKS device path (e.g., /dev/mapper/cryptroot), empty if not encrypted
	Subvolumes    []string // List of created Btrfs subvolumes (e.g., "@", "@home")
	MountedAt     []string // List of mount points (e.g., "/mnt", "/mnt/home", "/mnt/boot")
	ErrorDetail   string
}
