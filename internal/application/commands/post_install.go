package commands

// PostInstallCommand contains data for post-installation tasks
type PostInstallCommand struct {
	MountPoint         string // Root mount point
	Username           string // Standard user username
	UserEmail          string // User email for git config and SSH key (optional)
	RunPostBootScripts bool   // Whether to run post-boot scripts
	PlymouthTheme      string // Plymouth theme to install (optional)
	InstallDankLinux   bool   // Whether to write the Dank Linux flag file for first-boot auto-install
	TargetDisk         string // Target disk for bootloader hook (e.g. /dev/sda)
}
