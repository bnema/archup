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
	TargetDisk  string
	Success     bool
	Partitions  []*PartitionInfo
	ErrorDetail string
}
