package shell

import (
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// ConfigResult holds the result of shell configuration
type ConfigResult struct {
	ThemesApplied int
	ThemesFailed  []string
	Warnings      []string
}

// Configurator handles shell and bash configuration setup
type Configurator struct {
	fs      interfaces.FileSystem
	http    interfaces.HTTPClient
	chrExec interfaces.ChrootExecutor
	config  *config.Config
	logger  *logger.Logger
}

// NewConfigurator creates a new shell configurator
func NewConfigurator(
	fs interfaces.FileSystem,
	http interfaces.HTTPClient,
	chrExec interfaces.ChrootExecutor,
	cfg *config.Config,
	log *logger.Logger,
) *Configurator {
	return &Configurator{
		fs:      fs,
		http:    http,
		chrExec: chrExec,
		config:  cfg,
		logger:  log,
	}
}

// Configure is a no-op. Shell configuration is handled by cli-tools.sh at first boot.
func (c *Configurator) Configure(username, userHome string) (ConfigResult, error) {
	return ConfigResult{
		ThemesApplied: 0,
		ThemesFailed:  []string{},
		Warnings:      []string{},
	}, nil
}
