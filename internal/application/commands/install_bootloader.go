package commands

import "github.com/bnema/archup/internal/domain/bootloader"

// InstallBootloaderCommand contains data for bootloader installation
type InstallBootloaderCommand struct {
	MountPoint     string                    // Root mount point
	BootloaderType bootloader.BootloaderType // BootloaderTypeLimine or BootloaderTypeSystemdBoot
	TimeoutSeconds int                       // Boot menu timeout (0-600 seconds)
	Branding       string                    // Bootloader display name
}
