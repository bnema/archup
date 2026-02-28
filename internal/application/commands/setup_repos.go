package commands

import "github.com/bnema/archup/internal/domain/packages"

// SetupRepositoriesCommand contains data for repository configuration
type SetupRepositoriesCommand struct {
	MountPoint      string                 // Root mount point
	EnableMultilib  bool                   // Enable multilib repository
	AURHelper       packages.AURHelper     // AURHelperParu or AURHelperYay
	KernelVariant   packages.KernelVariant // Kernel variant for repo setup
	AdditionalRepos []string               // Additional repository URLs
}
