package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/ports"
)

// PartitionHandler handles disk partitioning
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

// Handle partitions the target disk
func (h *PartitionHandler) Handle(ctx context.Context, cmd commands.PartitionDiskCommand) (*dto.PartitionResult, error) {
	h.logger.Info("Starting disk partitioning", "disk", cmd.TargetDisk, "rootSizeGB", cmd.RootSizeGB)

	result := &dto.PartitionResult{
		TargetDisk:  cmd.TargetDisk,
		Success:     false,
		Partitions:  []*dto.PartitionInfo{},
		ErrorDetail: "",
	}

	// Create disk domain object
	diskObj, err := disk.NewDisk(cmd.TargetDisk, cmd.RootSizeGB+cmd.BootSizeGB+10)
	if err != nil {
		h.logger.Error("Failed to create disk object", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid disk configuration: %v", err)
		return result, err
	}

	// Create boot partition (EFI)
	bootPart, err := disk.NewPartition(
		cmd.TargetDisk+"1",
		cmd.BootSizeGB*1024, // Convert GB to MB
		disk.FilesystemFAT32,
		"/boot/efi",
		false, // Not encrypted
	)
	if err != nil {
		h.logger.Error("Failed to create boot partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create boot partition: %v", err)
		return result, err
	}

	if err := diskObj.AddPartition(bootPart); err != nil {
		h.logger.Error("Failed to add boot partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to add boot partition: %v", err)
		return result, err
	}

	// Create root partition
	rootPart, err := disk.NewPartition(
		cmd.TargetDisk+"2",
		cmd.RootSizeGB*1024, // Convert GB to MB
		cmd.FilesystemType,
		"/",
		cmd.EncryptionType != disk.EncryptionTypeNone,
	)
	if err != nil {
		h.logger.Error("Failed to create root partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create root partition: %v", err)
		return result, err
	}

	if err := diskObj.AddPartition(rootPart); err != nil {
		h.logger.Error("Failed to add root partition", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to add root partition: %v", err)
		return result, err
	}

	// Validate layout
	if err := diskObj.ValidateLayout(); err != nil {
		h.logger.Error("Disk layout validation failed", "error", err)
		result.ErrorDetail = fmt.Sprintf("Disk layout validation failed: %v", err)
		return result, err
	}

	// Execute partitioning (fdisk/parted commands would go here in real implementation)
	h.logger.Info("Disk layout validated, ready for partitioning", "disk", cmd.TargetDisk)

	// Build result with partition info
	for _, part := range diskObj.Partitions() {
		result.Partitions = append(result.Partitions, &dto.PartitionInfo{
			Device:     part.Device(),
			SizeGB:     part.SizeGB(),
			Filesystem: part.Filesystem().String(),
			MountPoint: part.MountPoint(),
			Encrypted:  part.IsEncrypted(),
		})
	}

	result.Success = true
	h.logger.Info("Disk partitioning completed successfully")
	return result, nil
}
