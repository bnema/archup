package disk

import (
	"testing"
)

// TestDeterminePartitionPath tests partition path determination
func TestDeterminePartitionPath(t *testing.T) {
	tests := []struct {
		name            string
		diskPath        string
		partitionNumber int
		expected        string
		shouldErr       bool
	}{
		// SATA/SSD cases
		{"SATA partition 1", "/dev/sda", 1, "/dev/sda1", false},
		{"SATA partition 2", "/dev/sda", 2, "/dev/sda2", false},
		{"SDB partition 1", "/dev/sdb", 1, "/dev/sdb1", false},
		// NVMe cases
		{"NVMe partition 1", "/dev/nvme0n1", 1, "/dev/nvme0n1p1", false},
		{"NVMe partition 2", "/dev/nvme0n1", 2, "/dev/nvme0n1p2", false},
		{"NVMe 1 partition 1", "/dev/nvme1n1", 1, "/dev/nvme1n1p1", false},
		// MMC cases
		{"MMC partition 1", "/dev/mmcblk0", 1, "/dev/mmcblk0p1", false},
		{"MMC partition 2", "/dev/mmcblk0", 2, "/dev/mmcblk0p2", false},
		// Virtio cases
		{"Virtio partition 1", "/dev/vda", 1, "/dev/vda1", false},
		{"Virtio partition 2", "/dev/vda", 2, "/dev/vda2", false},
		// Error cases
		{"empty disk path", "", 1, "", true},
		{"partition number 0", "/dev/sda", 0, "", true},
		{"partition number negative", "/dev/sda", -1, "", true},
		{"partition number too large", "/dev/sda", 129, "", true},
		{"invalid disk path", "sda", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeterminePartitionPath(tt.diskPath, tt.partitionNumber)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if !tt.shouldErr && result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestValidateDiskPath tests disk path validation
func TestValidateDiskPath(t *testing.T) {
	tests := []struct {
		name      string
		diskPath  string
		shouldErr bool
	}{
		{"valid SATA", "/dev/sda", false},
		{"valid SDB", "/dev/sdb", false},
		{"valid NVMe", "/dev/nvme0n1", false},
		{"valid MMC", "/dev/mmcblk0", false},
		{"valid Virtio", "/dev/vda", false},
		{"valid IDE", "/dev/hda", false},
		{"empty", "", true},
		{"no /dev/ prefix", "sda", true},
		{"invalid prefix", "/dev/xda", true},
		{"too long", "/dev/" + string(make([]byte, 60)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiskPath(tt.diskPath)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

// TestValidatePartitionPath tests partition path validation
func TestValidatePartitionPath(t *testing.T) {
	tests := []struct {
		name          string
		partitionPath string
		shouldErr     bool
	}{
		{"valid SATA partition", "/dev/sda1", false},
		{"valid NVMe partition", "/dev/nvme0n1p1", false},
		{"valid MMC partition", "/dev/mmcblk0p1", false},
		{"valid Virtio partition", "/dev/vda1", false},
		{"empty", "", true},
		{"no /dev/ prefix", "sda1", true},
		{"too long", "/dev/" + string(make([]byte, 60)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePartitionPath(tt.partitionPath)
			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}
		})
	}
}

// TestIsDiskPath tests disk vs partition detection
func TestIsDiskPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Disks
		{"SATA disk", "/dev/sda", true},
		{"SDB disk", "/dev/sdb", true},
		{"NVMe disk", "/dev/nvme0n1", true},
		{"MMC disk", "/dev/mmcblk0", true},
		{"Virtio disk", "/dev/vda", true},
		// Partitions
		{"SATA partition", "/dev/sda1", false},
		{"NVMe partition", "/dev/nvme0n1p1", false},
		{"MMC partition", "/dev/mmcblk0p1", false},
		{"Virtio partition", "/dev/vda1", false},
		// Edge cases
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDiskPath(tt.path)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestExtractDiskFromPartition tests extracting disk path from partition
func TestExtractDiskFromPartition(t *testing.T) {
	tests := []struct {
		name          string
		partitionPath string
		expected      string
		shouldErr     bool
	}{
		// SATA/SSD
		{"SATA partition 1", "/dev/sda1", "/dev/sda", false},
		{"SATA partition 10", "/dev/sda10", "/dev/sda", false},
		{"SDB partition 1", "/dev/sdb1", "/dev/sdb", false},
		// NVMe
		{"NVMe partition 1", "/dev/nvme0n1p1", "/dev/nvme0n1", false},
		{"NVMe partition 10", "/dev/nvme0n1p10", "/dev/nvme0n1", false},
		{"NVMe 1 partition 1", "/dev/nvme1n1p1", "/dev/nvme1n1", false},
		// MMC
		{"MMC partition 1", "/dev/mmcblk0p1", "/dev/mmcblk0", false},
		{"MMC partition 10", "/dev/mmcblk0p10", "/dev/mmcblk0", false},
		// Virtio
		{"Virtio partition 1", "/dev/vda1", "/dev/vda", false},
		{"Virtio partition 10", "/dev/vda10", "/dev/vda", false},
		// Already disk paths (should return as-is)
		{"SATA disk", "/dev/sda", "/dev/sda", false},
		{"NVMe disk", "/dev/nvme0n1", "/dev/nvme0n1", false},
		// Error cases
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractDiskFromPartition(tt.partitionPath)

			if (err != nil) != tt.shouldErr {
				t.Errorf("got error %v, expected error=%v", err, tt.shouldErr)
			}

			if !tt.shouldErr && result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDeterminePartitionPathConsistency tests consistency between DeterminePartitionPath and ExtractDiskFromPartition
func TestDeterminePartitionPathConsistency(t *testing.T) {
	disks := []string{
		"/dev/sda",
		"/dev/sdb",
		"/dev/nvme0n1",
		"/dev/nvme1n1",
		"/dev/mmcblk0",
		"/dev/vda",
	}

	for _, disk := range disks {
		for partNum := 1; partNum <= 5; partNum++ {
			// Determine partition path
			partPath, err := DeterminePartitionPath(disk, partNum)
			if err != nil {
				t.Fatalf("DeterminePartitionPath(%q, %d) failed: %v", disk, partNum, err)
			}

			// Extract disk from partition path
			extractedDisk, err := ExtractDiskFromPartition(partPath)
			if err != nil {
				t.Fatalf("ExtractDiskFromPartition(%q) failed: %v", partPath, err)
			}

			// Should match original disk
			if extractedDisk != disk {
				t.Errorf("Inconsistency for %q partition %d: determined %q, extracted %q",
					disk, partNum, partPath, extractedDisk)
			}
		}
	}
}

// TestValidateDiskPathErrors tests error messages
func TestValidateDiskPathErrors(t *testing.T) {
	tests := []struct {
		name      string
		diskPath  string
		errSubstr string
	}{
		{"empty", "", "cannot be empty"},
		{"no /dev/", "sda", "must start with /dev/"},
		{"invalid type", "/dev/xda", "unrecognized disk device type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDiskPath(tt.diskPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.errSubstr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

// TestValidatePartitionPathErrors tests error messages
func TestValidatePartitionPathErrors(t *testing.T) {
	tests := []struct {
		name          string
		partitionPath string
		errSubstr     string
	}{
		{"empty", "", "cannot be empty"},
		{"no /dev/", "sda1", "must start with /dev/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePartitionPath(tt.partitionPath)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.errSubstr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

// TestDeterminePartitionPathErrors tests error messages
func TestDeterminePartitionPathErrors(t *testing.T) {
	tests := []struct {
		name            string
		diskPath        string
		partitionNumber int
		errSubstr       string
	}{
		{"empty disk", "", 1, "cannot be empty"},
		{"invalid partition 0", "/dev/sda", 0, "between 1 and 128"},
		{"invalid partition negative", "/dev/sda", -1, "between 1 and 128"},
		{"invalid partition large", "/dev/sda", 129, "between 1 and 128"},
		{"no /dev/ prefix", "sda", 1, "must start with /dev/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DeterminePartitionPath(tt.diskPath, tt.partitionNumber)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.errSubstr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
			}
		})
	}
}
