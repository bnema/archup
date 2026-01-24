package commands

import "github.com/bnema/archup/internal/domain/packages"

// InstallBaseCommand contains data for base system installation
type InstallBaseCommand struct {
	TargetDisk       string                 // Target disk with partitions
	MountPoint       string                 // Root mount point (usually "/mnt")
	Packages         []string               // Additional packages to install (in addition to base)
	KernelVariant    packages.KernelVariant // KernelStable, KernelZen, KernelLTS, KernelHardened, KernelCachyOS
	IncludeMicrocode bool                   // Whether to install CPU microcode
}
