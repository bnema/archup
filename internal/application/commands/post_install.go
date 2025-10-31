package commands

// PostInstallCommand contains data for post-installation tasks
type PostInstallCommand struct {
	MountPoint       string   // Root mount point
	Username         string   // Standard user username
	RunPostBootScripts bool   // Whether to run post-boot scripts
	PlymouthTheme    string   // Plymouth theme to install (optional)
}
