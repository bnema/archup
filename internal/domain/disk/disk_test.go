package disk

import (
	"testing"
)

// Encryption Tests

func TestNewEncryptionConfig_None(t *testing.T) {
	enc, err := NewEncryptionConfig(EncryptionTypeNone, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if enc.IsEncrypted() {
		t.Error("expected not encrypted")
	}

	if enc.EncryptionType().String() != "None" {
		t.Errorf("expected type 'None'")
	}
}

func TestNewEncryptionConfig_LUKS(t *testing.T) {
	password := "test-password-12345"
	enc, err := NewEncryptionConfig(EncryptionTypeLUKS, password)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !enc.IsEncrypted() {
		t.Error("expected encrypted")
	}

	if enc.HasPassword() != true {
		t.Error("expected to have password")
	}
}

func TestNewEncryptionConfig_MissingPassword(t *testing.T) {
	_, err := NewEncryptionConfig(EncryptionTypeLUKS, "")

	if err == nil {
		t.Error("expected error for missing encryption password")
	}
}

func TestNewEncryptionConfig_WeakPassword(t *testing.T) {
	_, err := NewEncryptionConfig(EncryptionTypeLUKS, "weak")

	if err == nil {
		t.Error("expected error for weak password")
	}
}

func TestEncryptionType_String(t *testing.T) {
	tests := []struct {
		encType EncryptionType
		want    string
	}{
		{EncryptionTypeNone, "None"},
		{EncryptionTypeLUKS, "LUKS"},
		{EncryptionTypeLUKSLVM, "LUKS+LVM"},
	}

	for _, tt := range tests {
		if got := tt.encType.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

// Partition Tests

func TestNewPartition_Valid(t *testing.T) {
	partition, err := NewPartition("/dev/sda1", 50*1024, FilesystemExt4, "/", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if partition.Device() != "/dev/sda1" {
		t.Errorf("expected device /dev/sda1")
	}

	if !partition.IsRootPartition() {
		t.Error("expected to be root partition")
	}
}

func TestNewPartition_InvalidDevice(t *testing.T) {
	_, err := NewPartition("", 50*1024, FilesystemExt4, "/", false)

	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestNewPartition_InvalidSize(t *testing.T) {
	_, err := NewPartition("/dev/sda1", 0, FilesystemExt4, "/", false)

	if err == nil {
		t.Error("expected error for zero size")
	}
}

func TestNewPartition_InvalidMountPoint(t *testing.T) {
	_, err := NewPartition("/dev/sda1", 50*1024, FilesystemExt4, "not-absolute", false)

	if err == nil {
		t.Error("expected error for non-absolute mount point")
	}
}

func TestNewPartition_FAT32NotForRoot(t *testing.T) {
	_, err := NewPartition("/dev/sda1", 50*1024, FilesystemFAT32, "/", false)

	if err == nil {
		t.Error("expected error for FAT32 on root partition")
	}
}

func TestPartition_IsBootPartition(t *testing.T) {
	tests := []struct {
		name   string
		mount  string
		isBoot bool
	}{
		{"boot partition", "/boot", true},
		{"efi boot partition", "/boot/efi", true},
		{"root partition", "/", false},
		{"home partition", "/home", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := NewPartition("/dev/sda1", 10*1024, FilesystemExt4, tt.mount, false)
			if p.IsBootPartition() != tt.isBoot {
				t.Errorf("expected IsBootPartition=%v", tt.isBoot)
			}
		})
	}
}

func TestPartition_IsEFI(t *testing.T) {
	tests := []struct {
		name      string
		fs        FilesystemType
		mount     string
		isEFI     bool
		shouldErr bool
	}{
		{"EFI partition", FilesystemFAT32, "/boot/efi", true, false},
		{"EFI with ext4", FilesystemExt4, "/boot/efi", false, false},
		{"FAT32 non-EFI", FilesystemFAT32, "/boot", false, true}, // Should error
		{"ext4 root", FilesystemExt4, "/", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPartition("/dev/sda1", 1*1024, tt.fs, tt.mount, false)

			if tt.shouldErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if p.IsEFI() != tt.isEFI {
				t.Errorf("expected IsEFI=%v", tt.isEFI)
			}
		})
	}
}

// Disk Tests

func TestNewDisk_Valid(t *testing.T) {
	disk, err := NewDisk("/dev/sda", 500)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if disk.Device() != "/dev/sda" {
		t.Errorf("expected device /dev/sda")
	}

	if disk.SizeGB() != 500 {
		t.Errorf("expected size 500GB")
	}

	if disk.PartitionCount() != 0 {
		t.Error("expected no partitions initially")
	}
}

func TestNewDisk_InvalidDevice(t *testing.T) {
	_, err := NewDisk("", 500)

	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestNewDisk_InvalidSize(t *testing.T) {
	_, err := NewDisk("/dev/sda", 0)

	if err == nil {
		t.Error("expected error for zero size")
	}
}

func TestDisk_AddPartition(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)
	partition, _ := NewPartition("/dev/sda1", 20*1024, FilesystemExt4, "/", false)

	err := disk.AddPartition(partition)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if disk.PartitionCount() != 1 {
		t.Error("expected 1 partition")
	}
}

func TestDisk_AddPartition_InvalidDevice(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)
	partition, _ := NewPartition("/dev/sdb1", 20*1024, FilesystemExt4, "/", false)

	err := disk.AddPartition(partition)

	if err == nil {
		t.Error("expected error for partition on different disk")
	}
}

func TestDisk_AddPartition_DuplicateMountPoint(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)
	partition1, _ := NewPartition("/dev/sda1", 20*1024, FilesystemExt4, "/", false)
	partition2, _ := NewPartition("/dev/sda2", 20*1024, FilesystemExt4, "/", false)

	disk.AddPartition(partition1)
	err := disk.AddPartition(partition2)

	if err == nil {
		t.Error("expected error for duplicate mount point")
	}
}

func TestDisk_AddPartition_ExceedsCapacity(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 100)
	partition1, _ := NewPartition("/dev/sda1", 80*1024, FilesystemExt4, "/", false)
	partition2, _ := NewPartition("/dev/sda2", 50*1024, FilesystemExt4, "/home", false)

	disk.AddPartition(partition1)
	err := disk.AddPartition(partition2)

	if err == nil {
		t.Error("expected error when total partitions exceed disk size")
	}
}

func TestDisk_FindPartitionByMountPoint(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)
	partition, _ := NewPartition("/dev/sda1", 20*1024, FilesystemExt4, "/", false)
	disk.AddPartition(partition)

	found := disk.FindPartitionByMountPoint("/")
	if found == nil {
		t.Error("expected to find partition")
	}

	if found.Device() != "/dev/sda1" {
		t.Error("expected correct partition")
	}

	notFound := disk.FindPartitionByMountPoint("/home")
	if notFound != nil {
		t.Error("expected not to find partition")
	}
}

func TestDisk_FindPartitionByDevice(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)
	partition, _ := NewPartition("/dev/sda1", 20*1024, FilesystemExt4, "/", false)
	disk.AddPartition(partition)

	found := disk.FindPartitionByDevice("/dev/sda1")
	if found == nil {
		t.Error("expected to find partition")
	}

	notFound := disk.FindPartitionByDevice("/dev/sda2")
	if notFound != nil {
		t.Error("expected not to find partition")
	}
}

func TestDisk_HasRootPartition(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	if disk.HasRootPartition() {
		t.Error("expected no root partition initially")
	}

	partition, _ := NewPartition("/dev/sda1", 20*1024, FilesystemExt4, "/", false)
	disk.AddPartition(partition)

	if !disk.HasRootPartition() {
		t.Error("expected to have root partition")
	}
}

func TestDisk_ValidateLayout_Success(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	// Add EFI partition
	efi, _ := NewPartition("/dev/sda1", 512, FilesystemFAT32, "/boot/efi", false)
	disk.AddPartition(efi)

	// Add root partition
	root, _ := NewPartition("/dev/sda2", 50*1024, FilesystemExt4, "/", false)
	disk.AddPartition(root)

	err := disk.ValidateLayout()
	if err != nil {
		t.Errorf("expected valid layout, got error: %v", err)
	}
}

func TestDisk_ValidateLayout_MissingRoot(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	efi, _ := NewPartition("/dev/sda1", 512, FilesystemFAT32, "/boot/efi", false)
	disk.AddPartition(efi)

	err := disk.ValidateLayout()
	if err == nil {
		t.Error("expected error for missing root partition")
	}
}

func TestDisk_ValidateLayout_MissingEFI(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	root, _ := NewPartition("/dev/sda1", 50*1024, FilesystemExt4, "/", false)
	disk.AddPartition(root)

	err := disk.ValidateLayout()
	if err == nil {
		t.Error("expected error for missing EFI partition")
	}
}

func TestDisk_ValidateLayout_RootTooSmall(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	efi, _ := NewPartition("/dev/sda1", 512, FilesystemFAT32, "/boot/efi", false)
	disk.AddPartition(efi)

	root, _ := NewPartition("/dev/sda2", 10*1024, FilesystemExt4, "/", false)
	disk.AddPartition(root)

	err := disk.ValidateLayout()
	if err == nil {
		t.Error("expected error for root partition too small")
	}
}

func TestDisk_ValidateLayout_EFIWrongFilesystem(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	efi, _ := NewPartition("/dev/sda1", 512, FilesystemExt4, "/boot/efi", false)
	disk.AddPartition(efi)

	root, _ := NewPartition("/dev/sda2", 50*1024, FilesystemExt4, "/", false)
	disk.AddPartition(root)

	err := disk.ValidateLayout()
	if err == nil {
		t.Error("expected error for EFI partition not FAT32")
	}
}

func TestDisk_CalculateSpaces(t *testing.T) {
	disk, _ := NewDisk("/dev/sda", 500)

	efi, _ := NewPartition("/dev/sda1", 1*1024, FilesystemFAT32, "/boot/efi", false)
	disk.AddPartition(efi)

	root, _ := NewPartition("/dev/sda2", 100*1024, FilesystemExt4, "/", false)
	disk.AddPartition(root)

	used := disk.CalculateUsedSpace()
	if used != 101 { // 1GB + 100GB
		t.Errorf("expected 101GB used, got %dGB", used)
	}

	free := disk.CalculateFreeSpace()
	if free != 399 { // 500GB - 101GB
		t.Errorf("expected 399GB free, got %dGB", free)
	}
}

func TestIsPartitionOfDisk(t *testing.T) {
	tests := []struct {
		disk      string
		partition string
		expected  bool
	}{
		{"/dev/sda", "/dev/sda1", true},
		{"/dev/sda", "/dev/sda10", true},
		{"/dev/sda", "/dev/sdb1", false},
		{"/dev/sda", "/dev/sdaa1", false},
		{"/dev/nvme0n1", "/dev/nvme0n1p1", true},
		{"/dev/nvme0n1", "/dev/nvme0n1p10", true},
	}

	for _, tt := range tests {
		result := isPartitionOfDisk(tt.disk, tt.partition)
		if result != tt.expected {
			t.Errorf("isPartitionOfDisk(%s, %s): expected %v, got %v",
				tt.disk, tt.partition, tt.expected, result)
		}
	}
}

func TestFilesystemType_String(t *testing.T) {
	tests := []struct {
		fs   FilesystemType
		want string
	}{
		{FilesystemExt4, "ext4"},
		{FilesystemFAT32, "FAT32"},
		{FilesystemBtrfs, "Btrfs"},
	}

	for _, tt := range tests {
		if got := tt.fs.String(); got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}
