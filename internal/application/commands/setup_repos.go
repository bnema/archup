package commands

import "github.com/bnema/archup/internal/domain/packages"

// SetupRepositoriesCommand contains data for repository configuration
type SetupRepositoriesCommand struct {
	MountPoint      string // Root mount point
	EnableMultilib  bool   // Enable multilib repository
	EnableChaotic   bool   // Enable Chaotic-AUR repository
	AURHelper       packages.AURHelper // AURHelperParu or AURHelperYay
	AdditionalRepos []string // Additional repository URLs
}
