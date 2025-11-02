package shell

import (
	"fmt"
	"path/filepath"
	"strings"

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

// Configure sets up shell and CLI tool theming for a user
func (c *Configurator) Configure(username, userHome string) (ConfigResult, error) {
	result := ConfigResult{
		ThemesApplied: 0,
		ThemesFailed:  []string{},
		Warnings:      []string{},
	}

	// Create archup shell directory structure
	archupDefault := filepath.Join(userHome, ".local", "share", "archup", "default")
	archupDefaultBash := filepath.Join(archupDefault, "bash")

	if err := c.fs.MkdirAll(archupDefaultBash, 0755); err != nil {
		return result, fmt.Errorf("failed to create shell config directory: %w", err)
	}

	// Copy shell configuration files from templates
	if err := c.copyShellTemplates(archupDefault, archupDefaultBash); err != nil {
		return result, fmt.Errorf("failed to copy shell templates: %w", err)
	}

	// Create .bashrc
	bashrcContent, err := c.readTemplate(config.BashrcTemplate)
	if err != nil {
		return result, fmt.Errorf("failed to read bashrc template: %w", err)
	}

	bashrcPath := filepath.Join(userHome, ".bashrc")
	if err := c.fs.WriteFile(bashrcPath, bashrcContent, 0644); err != nil {
		return result, fmt.Errorf("failed to write .bashrc: %w", err)
	}

	// Set ownership to user
	if err := c.setShellOwnership(userHome, username); err != nil {
		return result, fmt.Errorf("failed to set ownership: %w", err)
	}

	// Clone bleu-theme repository
	relativeHome := strings.TrimPrefix(userHome, config.PathMnt)
	themesDir := filepath.Join(relativeHome, ".local", "share", "archup", "themes")
	bleuThemeDir := filepath.Join(themesDir, "bleu")
	currentDir := filepath.Join(relativeHome, ".local", "share", "archup", "current")

	cloneCmd := fmt.Sprintf("su - %s -c 'mkdir -p %s && git clone https://github.com/bnema/bleu-theme.git %s'",
		username, themesDir, bleuThemeDir)
	if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, cloneCmd); err != nil {
		result.Warnings = append(result.Warnings, "Failed to clone bleu-theme repository")
	}

	// Create current theme symlink
	symlinkCmd := fmt.Sprintf("su - %s -c 'mkdir -p %s && ln -snf %s %s/theme'",
		username, currentDir, bleuThemeDir, currentDir)
	if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, symlinkCmd); err != nil {
		result.Warnings = append(result.Warnings, "Failed to create theme symlink")
	}

	// Apply themes to CLI tools
	themeResult := c.applyThemes(username, userHome)
	result.ThemesApplied = themeResult.Applied
	result.ThemesFailed = themeResult.Failed
	result.Warnings = append(result.Warnings, themeResult.Warnings...)

	// Configure git delta
	gitCommands := []string{
		fmt.Sprintf("su - %s -c 'git config --global core.pager delta'", username),
		fmt.Sprintf("su - %s -c 'git config --global interactive.diffFilter \"delta --color-only\"'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.navigate true'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.light false'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.side-by-side true'", username),
		fmt.Sprintf("su - %s -c 'git config --global merge.conflictstyle diff3'", username),
		fmt.Sprintf("su - %s -c 'git config --global diff.colorMoved default'", username),
	}

	for _, gitCmd := range gitCommands {
		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, gitCmd); err != nil {
			result.Warnings = append(result.Warnings, "Failed to configure git delta")
			break
		}
	}

	return result, nil
}

// copyShellTemplates copies shell config templates to user directory
func (c *Configurator) copyShellTemplates(archupDefault, archupDefaultBash string) error {
	templates := map[string]string{
		config.ShellConfigTemplate:  filepath.Join(archupDefaultBash, "shell"),
		config.ShellInitTemplate:    filepath.Join(archupDefaultBash, "init"),
		config.ShellAliasesTemplate: filepath.Join(archupDefaultBash, "aliases"),
		config.ShellEnvsTemplate:    filepath.Join(archupDefaultBash, "envs"),
		config.ShellRcTemplate:      filepath.Join(archupDefaultBash, "rc"),
	}

	for template, dest := range templates {
		content, err := c.readTemplate(template)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", template, err)
		}

		if err := c.fs.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dest, err)
		}
	}

	return nil
}

// readTemplate reads a template file from install path
func (c *Configurator) readTemplate(filename string) ([]byte, error) {
	// Use DefaultInstallDir directly to match where bootstrap downloads files
	templatePath := filepath.Join(config.DefaultInstallDir, filename)
	return c.fs.ReadFile(templatePath)
}

// setShellOwnership sets ownership of shell config files to user
func (c *Configurator) setShellOwnership(userHome, username string) error {
	// Use chroot to run chown
	relativeHome := strings.TrimPrefix(userHome, config.PathMnt)
	chownCmd := fmt.Sprintf("chown -R %s:%s %s/.local %s/.bashrc",
		username, username, relativeHome, relativeHome)
	return c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, chownCmd)
}

