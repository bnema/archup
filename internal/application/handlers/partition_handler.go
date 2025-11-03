package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/ports"
)

// PartitionHandler handles disk partitioning, formatting, and mounting
type PartitionHandler struct {
	cmdExec ports.CommandExecutor
	logger  ports.Logger
}

// NewPartitionHandler creates a new partition handler
func NewPartitionHandler(cmdExec ports.CommandExecutor, logger ports.Logger) *PartitionHandler {
	return &PartitionHandler{
		cmdExec: cmdExec,
		logger:  logger,
	}
}

// Handle executes the complete disk partitioning workflow
func (h *PartitionHandler) Handle(ctx context.Context, cmd commands.PartitionDiskCommand) (*dto.PartitionResult, error) {
	h.logger.Info("Starting disk partitioning", "disk", cmd.TargetDisk)

	result := &dto.PartitionResult{
		TargetDisk:  cmd.TargetDisk,
		Success:     false,
		Partitions:  []*dto.PartitionInfo{},
		ErrorDetail: "",
	}

	// Step 1: Wipe disk if requested
	if cmd.WipeDisks {
		h.logger.Info("Wiping disk", "disk", cmd.TargetDisk)
		if err := h.wipeDisks(ctx, cmd.TargetDisk); err != nil {
			h.logger.Error("Failed to wipe disk", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to wipe disk: %v", err)
			return result, err
		}
	}

	// Step 2: Create GPT partitions (EFI + ROOT)
	h.logger.Info("Creating GPT partition table")
	efiPartition, rootPartition, err := h.createGPTPartitions(ctx, cmd.TargetDisk)
	if err != nil {
		h.logger.Error("Failed to create partitions", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create partitions: %v", err)
		return result, err
	}

	result.EFIPartition = efiPartition
	result.RootPartition = rootPartition

	// Step 3: Format EFI partition as FAT32
	h.logger.Info("Formatting EFI partition", "partition", efiPartition)
	if err := h.formatEFIPartition(ctx, efiPartition); err != nil {
		h.logger.Error("Failed to format EFI partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to format EFI partition: %v", err)
		return result, err
	}

	// Step 4: Handle root partition (with optional encryption)
	var rootDevice string
	if cmd.EncryptionType != disk.EncryptionTypeNone {
		h.logger.Info("Setting up LUKS encryption")
		cryptDevice, err := h.setupLUKSEncryption(ctx, rootPartition, cmd.EncryptionPassword)
		if err != nil {
			h.logger.Error("Failed to setup LUKS encryption", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to setup encryption: %v", err)
			return result, err
		}
		result.CryptDevice = cryptDevice
		rootDevice = cryptDevice
	} else {
		rootDevice = rootPartition
	}

	// Step 5: Format root partition as Btrfs
	h.logger.Info("Formatting root partition as Btrfs", "device", rootDevice)
	if err := h.formatRootPartition(ctx, rootDevice); err != nil {
		h.logger.Error("Failed to format root partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to format root partition: %v", err)
		return result, err
	}

	// Step 6: Create Btrfs subvolumes
	h.logger.Info("Creating Btrfs subvolumes")
	layout, err := disk.NewStandardBtrfsLayout()
	if err != nil {
		h.logger.Error("Failed to create Btrfs layout", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create Btrfs layout: %v", err)
		return result, err
	}

	subvolumes, err := h.createBtrfsSubvolumes(ctx, rootDevice, layout)
	if err != nil {
		h.logger.Error("Failed to create Btrfs subvolumes", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create Btrfs subvolumes: %v", err)
		return result, err
	}
	result.Subvolumes = subvolumes

	// Step 7: Mount filesystems
	h.logger.Info("Mounting filesystems")
	mounts, err := h.mountFilesystems(ctx, efiPartition, rootDevice, layout)
	if err != nil {
		h.logger.Error("Failed to mount filesystems", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to mount filesystems: %v", err)
		return result, err
	}
	result.MountedAt = mounts

	// Build partition info for result
	result.Partitions = []*dto.PartitionInfo{
		{
			Device:     efiPartition,
			SizeGB:     cmd.BootSizeGB,
			Filesystem: "FAT32",
			MountPoint: "/boot",
			Encrypted:  false,
		},
		{
			Device:     rootPartition,
			SizeGB:     cmd.RootSizeGB,
			Filesystem: "Btrfs",
			MountPoint: "/",
			Encrypted:  cmd.EncryptionType != disk.EncryptionTypeNone,
		},
	}

	result.Success = true
	h.logger.Info("Disk partitioning completed successfully")
	return result, nil
}

// wipeDisks wipes filesystem signatures from disk
func (h *PartitionHandler) wipeDisks(ctx context.Context, diskPath string) error {
	h.logger.Info("Wiping filesystem signatures", "disk", diskPath)

	// Validate disk path
	if err := disk.ValidateDiskPath(diskPath); err != nil {
		return fmt.Errorf("invalid disk path: %w", err)
	}

	// wipefs -af removes all signatures
	if _, err := h.cmdExec.Execute(ctx, "wipefs", "-af", diskPath); err != nil {
		return fmt.Errorf("wipefs failed: %w", err)
	}

	// sgdisk --zap-all removes GPT and MBR
	if _, err := h.cmdExec.Execute(ctx, "sgdisk", "--zap-all", diskPath); err != nil {
		return fmt.Errorf("sgdisk zap failed: %w", err)
	}

	h.logger.Info("Disk wiped successfully", "disk", diskPath)
	return nil
}

// createGPTPartitions creates GPT partition table with EFI (512MB) and ROOT partitions
func (h *PartitionHandler) createGPTPartitions(ctx context.Context, diskPath string) (string, string, error) {
	h.logger.Info("Creating GPT partition table", "disk", diskPath)

	// Validate disk path
	if err := disk.ValidateDiskPath(diskPath); err != nil {
		return "", "", fmt.Errorf("invalid disk path: %w", err)
	}

	// Create partitions using sgdisk
	// Partition 1: 512MB EFI (type ef00)
	// Partition 2: Remaining space ROOT (type 8300)
	if _, err := h.cmdExec.Execute(ctx, "sgdisk",
		"--clear",
		"--new=1:0:+512M", "--typecode=1:ef00", "--change-name=1:EFI",
		"--new=2:0:0", "--typecode=2:8300", "--change-name=2:ROOT",
		diskPath); err != nil {
		return "", "", fmt.Errorf("sgdisk partition creation failed: %w", err)
	}

	// Inform kernel of partition table changes
	if _, err := h.cmdExec.Execute(ctx, "partprobe", diskPath); err != nil {
		h.logger.Warn("partprobe failed (not critical)", "error", err)
	}

	// Determine partition paths based on disk type
	efiPartition, err := disk.DeterminePartitionPath(diskPath, 1)
	if err != nil {
		return "", "", fmt.Errorf("failed to determine EFI partition path: %w", err)
	}

	rootPartition, err := disk.DeterminePartitionPath(diskPath, 2)
	if err != nil {
		return "", "", fmt.Errorf("failed to determine ROOT partition path: %w", err)
	}

	h.logger.Info("Partitions created", "efi", efiPartition, "root", rootPartition)
	return efiPartition, rootPartition, nil
}

// formatEFIPartition formats EFI partition as FAT32
func (h *PartitionHandler) formatEFIPartition(ctx context.Context, partitionPath string) error {
	h.logger.Info("Formatting EFI partition", "partition", partitionPath)

	// Validate partition path
	if err := disk.ValidatePartitionPath(partitionPath); err != nil {
		return fmt.Errorf("invalid partition path: %w", err)
	}

	// Wipe partition first
	if _, err := h.cmdExec.Execute(ctx, "wipefs", "-af", partitionPath); err != nil {
		return fmt.Errorf("wipefs EFI partition failed: %w", err)
	}

	// Format as FAT32 with label "EFI"
	if _, err := h.cmdExec.Execute(ctx, "mkfs.fat", "-F32", "-n", "EFI", partitionPath); err != nil {
		return fmt.Errorf("mkfs.fat failed: %w", err)
	}

	h.logger.Info("EFI partition formatted successfully")
	return nil
}

// setupLUKSEncryption sets up LUKS2 encryption on root partition
func (h *PartitionHandler) setupLUKSEncryption(ctx context.Context, partitionPath, password string) (string, error) {
	h.logger.Info("Setting up LUKS encryption", "partition", partitionPath)

	// Validate partition path
	if err := disk.ValidatePartitionPath(partitionPath); err != nil {
		return "", fmt.Errorf("invalid partition path: %w", err)
	}

	// Validate encryption password
	encConfig, err := disk.NewEncryptionConfig(disk.EncryptionTypeLUKS, password)
	if err != nil {
		return "", fmt.Errorf("invalid encryption configuration: %w", err)
	}

	if !encConfig.IsEncrypted() {
		return "", fmt.Errorf("encryption config invalid")
	}

	// Wipe partition first
	if _, err := h.cmdExec.Execute(ctx, "wipefs", "-af", partitionPath); err != nil {
		return "", fmt.Errorf("wipefs root partition failed: %w", err)
	}

	// Create LUKS2 container with Argon2id
	// Note: Using printf to pipe password securely
	luksFormatCmd := fmt.Sprintf("printf '%%s' '%s' | cryptsetup luksFormat --type luks2 --batch-mode --pbkdf argon2id --iter-time 2000 --label ARCHUP_LUKS --key-file=- %s",
		password, partitionPath)

	if _, err := h.cmdExec.Execute(ctx, "sh", "-c", luksFormatCmd); err != nil {
		return "", fmt.Errorf("luksFormat failed: %w", err)
	}

	// Open LUKS container
	cryptDevice := "/dev/mapper/cryptroot"
	luksOpenCmd := fmt.Sprintf("printf '%%s' '%s' | cryptsetup open --key-file=- %s cryptroot",
		password, partitionPath)

	if _, err := h.cmdExec.Execute(ctx, "sh", "-c", luksOpenCmd); err != nil {
		return "", fmt.Errorf("cryptsetup open failed: %w", err)
	}

	h.logger.Info("LUKS encryption setup successfully", "cryptDevice", cryptDevice)
	return cryptDevice, nil
}

// formatRootPartition formats root partition (or encrypted device) as Btrfs
func (h *PartitionHandler) formatRootPartition(ctx context.Context, devicePath string) error {
	h.logger.Info("Formatting root partition as Btrfs", "device", devicePath)

	// Format as Btrfs with label "ROOT"
	if _, err := h.cmdExec.Execute(ctx, "mkfs.btrfs", "-f", "-L", "ROOT", devicePath); err != nil {
		return fmt.Errorf("mkfs.btrfs failed: %w", err)
	}

	h.logger.Info("Root partition formatted successfully")
	return nil
}

// createBtrfsSubvolumes creates Btrfs subvolumes according to layout
func (h *PartitionHandler) createBtrfsSubvolumes(ctx context.Context, devicePath string, layout *disk.BtrfsLayout) ([]string, error) {
	h.logger.Info("Creating Btrfs subvolumes", "device", devicePath)

	// Validate layout
	if err := layout.ValidateLayout(); err != nil {
		return nil, fmt.Errorf("invalid Btrfs layout: %w", err)
	}

	// Mount temporarily to /mnt
	if _, err := h.cmdExec.Execute(ctx, "mount", devicePath, "/mnt"); err != nil {
		return nil, fmt.Errorf("temporary mount failed: %w", err)
	}

	// Ensure unmount on exit
	defer func() {
		if _, err := h.cmdExec.Execute(ctx, "umount", "/mnt"); err != nil {
			h.logger.Warn("Failed to unmount temporary mount", "error", err)
		}
	}()

	subvolumes := []string{}

	// Create each subvolume
	for _, sv := range layout.Subvolumes() {
		subvolPath := "/mnt/" + sv.Name()
		h.logger.Info("Creating subvolume", "name", sv.Name())

		if _, err := h.cmdExec.Execute(ctx, "btrfs", "subvolume", "create", subvolPath); err != nil {
			return nil, fmt.Errorf("failed to create subvolume %s: %w", sv.Name(), err)
		}

		subvolumes = append(subvolumes, sv.Name())
	}

	h.logger.Info("Btrfs subvolumes created successfully", "count", len(subvolumes))
	return subvolumes, nil
}

// mountFilesystems mounts all filesystems to /mnt with proper options
func (h *PartitionHandler) mountFilesystems(ctx context.Context, efiPartition, rootDevice string, layout *disk.BtrfsLayout) ([]string, error) {
	h.logger.Info("Mounting filesystems")

	mounts := []string{}

	// Get root subvolume
	rootSubvolume := layout.GetRootSubvolume()
	if rootSubvolume == nil {
		return nil, fmt.Errorf("no root subvolume found in layout")
	}

	// Mount @ subvolume as root with Btrfs options
	rootMountOpts, err := disk.NewBtrfsMountOptions(rootSubvolume.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create mount options: %w", err)
	}

	if _, err := h.cmdExec.Execute(ctx, "mount", "-o", rootMountOpts.ToString(), rootDevice, "/mnt"); err != nil {
		return nil, fmt.Errorf("failed to mount @ subvolume: %w", err)
	}
	mounts = append(mounts, "/mnt")

	// Mount @home subvolume if it exists
	homeSubvolume := layout.FindSubvolumeByName("@home")
	if homeSubvolume != nil {
		// Create /mnt/home directory
		if _, err := h.cmdExec.Execute(ctx, "mkdir", "-p", "/mnt/home"); err != nil {
			return nil, fmt.Errorf("failed to create /mnt/home: %w", err)
		}

		homeMountOpts, err := disk.NewBtrfsMountOptions(homeSubvolume.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to create home mount options: %w", err)
		}

		if _, err := h.cmdExec.Execute(ctx, "mount", "-o", homeMountOpts.ToString(), rootDevice, "/mnt/home"); err != nil {
			return nil, fmt.Errorf("failed to mount @home subvolume: %w", err)
		}
		mounts = append(mounts, "/mnt/home")
	}

	// Mount EFI partition to /mnt/boot
	if _, err := h.cmdExec.Execute(ctx, "mkdir", "-p", "/mnt/boot"); err != nil {
		return nil, fmt.Errorf("failed to create /mnt/boot: %w", err)
	}

	if _, err := h.cmdExec.Execute(ctx, "mount", efiPartition, "/mnt/boot"); err != nil {
		return nil, fmt.Errorf("failed to mount EFI partition: %w", err)
	}
	mounts = append(mounts, "/mnt/boot")

	h.logger.Info("All filesystems mounted successfully", "mounts", mounts)
	return mounts, nil
}

// Rollback unmounts filesystems and closes LUKS device
func (h *PartitionHandler) Rollback(ctx context.Context, result *dto.PartitionResult) error {
	h.logger.Warn("Rolling back partitioning changes")

	// Unmount in reverse order
	for i := len(result.MountedAt) - 1; i >= 0; i-- {
		mountPoint := result.MountedAt[i]
		h.logger.Info("Unmounting", "mount", mountPoint)
		if _, err := h.cmdExec.Execute(ctx, "umount", mountPoint); err != nil {
			h.logger.Warn("Failed to unmount", "mount", mountPoint, "error", err)
		}
	}

	// Close LUKS device if it exists
	if result.CryptDevice != "" {
		h.logger.Info("Closing LUKS device", "device", result.CryptDevice)
		if _, err := h.cmdExec.Execute(ctx, "cryptsetup", "close", "cryptroot"); err != nil {
			h.logger.Warn("Failed to close LUKS device", "error", err)
		}
	}

	h.logger.Info("Rollback completed")
	return nil
}
