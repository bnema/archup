package phases

import (
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
)

// PartitioningPhase handles disk partitioning, formatting, and mounting
type PartitioningPhase struct {
	*BasePhase
}

// NewPartitioningPhase creates a new partitioning phase
func NewPartitioningPhase(cfg *config.Config, log *logger.Logger) *PartitioningPhase {
	return &PartitioningPhase{
		BasePhase: NewBasePhase("partitioning", "Disk Partitioning", cfg, log),
	}
}

// PreCheck validates partitioning prerequisites
func (p *PartitioningPhase) PreCheck() error {
	switch {
	case p.config.TargetDisk == "":
		return fmt.Errorf("target disk not selected")
	}

	return nil
}

// Execute runs the partitioning phase
func (p *PartitioningPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting partitioning...", 1, 5)

	// Step 1: Wipe and partition
	switch err := p.createPartitions(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Formatting partitions...", 2, 5)

	// Step 2: Format partitions
	switch err := p.formatPartitions(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Creating btrfs subvolumes...", 3, 5)

	// Step 3: Create subvolumes
	switch err := p.createSubvolumes(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Mounting filesystems...", 4, 5)

	// Step 4: Mount
	switch err := p.mountFilesystems(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Partitioning complete", 5, 5)
	p.SendComplete(progressChan, "Disk prepared successfully")

	return PhaseResult{Success: true, Message: "Partitioning complete"}
}

// createPartitions wipes disk and creates GPT partitions
func (p *PartitioningPhase) createPartitions(progressChan chan<- ProgressUpdate) error {
	disk := p.config.TargetDisk
	p.SendOutput(progressChan, fmt.Sprintf("Wiping disk %s...", disk))

	// Wipe filesystem signatures
	result := p.logger.ExecCommand("wipefs", "-af", disk)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("wipefs failed: %w", result.Error)
	}

	// Zap GPT and MBR
	result = p.logger.ExecCommand("sgdisk", "--zap-all", disk)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("sgdisk zap failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "Creating GPT partition table...")

	// Create partitions: 512MB EFI + remaining ROOT
	result = p.logger.ExecCommand("sgdisk",
		"--clear",
		"--new=1:0:+512M", "--typecode=1:ef00", "--change-name=1:EFI",
		"--new=2:0:0", "--typecode=2:8300", "--change-name=2:ROOT",
		disk)

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("sgdisk partition creation failed: %w", result.Error)
	}

	// Inform kernel of partition table changes
	result = p.logger.ExecCommand("partprobe", disk)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("partprobe failed: %w", result.Error)
	}

	// Determine partition naming (nvme vs sata/ssd)
	var efiPart, rootPart string
	switch {
	case strings.Contains(disk, "nvme"):
		efiPart = fmt.Sprintf("%sp1", disk)
		rootPart = fmt.Sprintf("%sp2", disk)
	default:
		efiPart = fmt.Sprintf("%s1", disk)
		rootPart = fmt.Sprintf("%s2", disk)
	}

	p.config.EFIPartition = efiPart
	p.config.RootPartition = rootPart

	p.SendOutput(progressChan, fmt.Sprintf("[OK] EFI: %s (512MB)", efiPart))
	p.SendOutput(progressChan, fmt.Sprintf("[OK] Root: %s", rootPart))

	return nil
}

// formatPartitions formats EFI and ROOT partitions
func (p *PartitioningPhase) formatPartitions(progressChan chan<- ProgressUpdate) error {
	// Format EFI as FAT32
	p.SendOutput(progressChan, "Formatting EFI partition...")

	result := p.logger.ExecCommand("wipefs", "-af", p.config.EFIPartition)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("wipefs EFI failed: %w", result.Error)
	}

	result = p.logger.ExecCommand("mkfs.fat", "-F32", "-n", "EFI", p.config.EFIPartition)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mkfs.fat failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] EFI formatted as FAT32")

	// Handle root partition with optional encryption
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		return p.formatEncryptedRoot(progressChan)
	default:
		return p.formatPlainRoot(progressChan)
	}
}

// formatEncryptedRoot sets up LUKS encryption on root partition
func (p *PartitioningPhase) formatEncryptedRoot(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Setting up LUKS encryption...")

	// Use user password for encryption (as per preflight)
	password := p.config.UserPassword
	switch {
	case password == "":
		return fmt.Errorf("password required for encryption")
	}

	// Wipe root partition
	result := p.logger.ExecCommand("wipefs", "-af", p.config.RootPartition)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("wipefs root failed: %w", result.Error)
	}

	// Create LUKS container with Argon2id
	p.SendOutput(progressChan, "Creating LUKS container (this may take a moment)...")

	// Note: In real implementation, pipe password securely via stdin
	// For now, documenting the approach
	result = system.RunSimple("sh", "-c",
		fmt.Sprintf("printf '%%s' '%s' | cryptsetup luksFormat --type luks2 --batch-mode --pbkdf argon2id --iter-time 2000 --label ARCHUP_LUKS --key-file - %s",
			password, p.config.RootPartition))

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("luksFormat failed: %w", result.Error)
	}

	// Open LUKS container
	result = system.RunSimple("sh", "-c",
		fmt.Sprintf("printf '%%s' '%s' | cryptsetup open --key-file=- %s cryptroot",
			password, p.config.RootPartition))

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("cryptsetup open failed: %w", result.Error)
	}

	p.config.CryptDevice = "/dev/mapper/cryptroot"
	p.SendOutput(progressChan, "[OK] LUKS container created and opened")

	// Format encrypted device with btrfs
	result = p.logger.ExecCommand("mkfs.btrfs", "-f", "-L", "ROOT", p.config.CryptDevice)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mkfs.btrfs on encrypted device failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Encrypted partition formatted as btrfs")
	return nil
}

// formatPlainRoot formats root partition without encryption
func (p *PartitioningPhase) formatPlainRoot(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Formatting root partition...")

	result := p.logger.ExecCommand("mkfs.btrfs", "-f", "-L", "ROOT", p.config.RootPartition)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mkfs.btrfs failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Root formatted as btrfs")
	return nil
}

// createSubvolumes creates btrfs subvolumes
func (p *PartitioningPhase) createSubvolumes(progressChan chan<- ProgressUpdate) error {
	// Determine device to mount
	var device string
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		device = p.config.CryptDevice
	default:
		device = p.config.RootPartition
	}

	// Mount temporarily
	result := p.logger.ExecCommand("mount", device, "/mnt")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("temporary mount failed: %w", result.Error)
	}

	// Create @ subvolume
	result = p.logger.ExecCommand("btrfs", "subvolume", "create", "/mnt/@")
	switch {
	case result.ExitCode != 0:
		system.RunSimple("umount", "/mnt") // Cleanup
		return fmt.Errorf("@ subvolume creation failed: %w", result.Error)
	}

	// Create @home subvolume
	result = p.logger.ExecCommand("btrfs", "subvolume", "create", "/mnt/@home")
	switch {
	case result.ExitCode != 0:
		system.RunSimple("umount", "/mnt") // Cleanup
		return fmt.Errorf("@home subvolume creation failed: %w", result.Error)
	}

	// Unmount
	result = p.logger.ExecCommand("umount", "/mnt")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("unmount after subvolume creation failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Btrfs subvolumes created: @ and @home")
	return nil
}

// mountFilesystems mounts all filesystems to /mnt
func (p *PartitioningPhase) mountFilesystems(progressChan chan<- ProgressUpdate) error {
	// Determine device
	var device string
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		device = p.config.CryptDevice
	default:
		device = p.config.RootPartition
	}

	// Mount @ subvolume as root
	result := p.logger.ExecCommand("mount", "-o", "noatime,compress=zstd,subvol=@", device, "/mnt")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mount @ failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Mounted @ to /mnt")

	// Create home directory
	result = p.logger.ExecCommand("mkdir", "-p", "/mnt/home")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mkdir /mnt/home failed: %w", result.Error)
	}

	// Mount @home subvolume
	result = p.logger.ExecCommand("mount", "-o", "noatime,compress=zstd,subvol=@home", device, "/mnt/home")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mount @home failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Mounted @home to /mnt/home")

	// Create and mount EFI
	result = p.logger.ExecCommand("mkdir", "-p", "/mnt/boot")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mkdir /mnt/boot failed: %w", result.Error)
	}

	result = p.logger.ExecCommand("mount", p.config.EFIPartition, "/mnt/boot")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("mount EFI failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Mounted EFI to /mnt/boot")
	return nil
}

// PostCheck validates mounts
func (p *PartitioningPhase) PostCheck() error {
	// Verify /mnt is mounted
	result := system.RunSimple("mountpoint", "-q", "/mnt")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("/mnt is not mounted")
	}

	return p.config.Save()
}

// Rollback unmounts filesystems
func (p *PartitioningPhase) Rollback() error {
	// Unmount in reverse order
	system.RunSimple("umount", "/mnt/boot")
	system.RunSimple("umount", "/mnt/home")
	system.RunSimple("umount", "/mnt")

	// Close LUKS if encrypted
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		switch {
		case p.config.CryptDevice != "":
			system.RunSimple("cryptsetup", "close", "cryptroot")
		}
	}

	return nil
}

// CanSkip returns false - partitioning cannot be skipped
func (p *PartitioningPhase) CanSkip() bool {
	return false
}
