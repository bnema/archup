package commands

import (
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/disk"
	"github.com/bnema/archup/internal/domain/packages"
)

// InstallBootloaderCommand contains data for bootloader installation
type InstallBootloaderCommand struct {
	MountPoint     string                    // Root mount point
	BootloaderType bootloader.BootloaderType // BootloaderTypeLimine
	TimeoutSeconds int                       // Boot menu timeout (0-600 seconds)
	Branding       string                    // Bootloader display name
	KernelVariant  packages.KernelVariant    // KernelStable, KernelZen, KernelLTS, KernelHardened, KernelCachyOS
	RootPartition  string                    // Root partition device path
	EncryptionType disk.EncryptionType       // EncryptionTypeNone, EncryptionTypeLUKS, EncryptionTypeLUKSLVM
	EFIPartition   string                    // EFI partition device path
	TargetDisk     string                    // Target disk device path
}
