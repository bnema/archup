package shell

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/config"
)

// ThemeResult holds the result of theme application
type ThemeResult struct {
	Applied  int
	Failed   []string
	Warnings []string
}

// applyThemes orchestrates centralized theme symlink setup for all CLI tools
func (c *Configurator) applyThemes(username, userHome string) ThemeResult {
	result := ThemeResult{
		Applied:  0,
		Failed:   []string{},
		Warnings: []string{},
	}

	// Strip /mnt prefix for use in chroot commands
	relativeHome := strings.TrimPrefix(userHome, config.PathMnt)

	// Step 1: Setup central theme directory structure
	if err := c.setupCentralThemeDirectory(username, relativeHome); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to setup central theme directory: %v", err))
		return result
	}

	// Step 2: Create central symlinks pointing to active theme (bleu)
	if err := c.createCentralSymlinks(username, relativeHome, "bleu"); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to create central symlinks: %v", err))
		return result
	}

	// Step 3: Apply each tool's theme configuration
	tools := []struct {
		name  string
		apply func(string, string) error
	}{
		{"starship", c.applyStarshipTheme},
		{"eza", c.applyEzaTheme},
		{"bat", c.applyBatTheme},
		{"btop", c.applyBtopTheme},
		{"yazi", c.applyYaziTheme},
	}

	for _, tool := range tools {
		if err := tool.apply(username, relativeHome); err != nil {
			result.Failed = append(result.Failed, tool.name)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to configure %s: %v", tool.name, err))
		} else {
			result.Applied++
		}
	}

	return result
}

// setupCentralThemeDirectory creates ~/.local/share/archup/themes/current/
func (c *Configurator) setupCentralThemeDirectory(username, relativeHome string) error {
	centralThemeDir := filepath.Join(relativeHome, ".local", "share", "archup", "themes", "current")

	mkdirCmd := fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, centralThemeDir)
	if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create central theme directory: %w", err)
	}

	return nil
}

// createCentralSymlinks creates symlinks in themes/current/ pointing to active theme files
func (c *Configurator) createCentralSymlinks(username, relativeHome, activeTheme string) error {
	themesDir := filepath.Join(relativeHome, ".local", "share", "archup", "themes")
	centralDir := filepath.Join(themesDir, "current")
	themeSource := filepath.Join(themesDir, activeTheme)

	// Map of symlink name -> target path
	symlinks := map[string]string{
		"starship.toml": filepath.Join(themeSource, "starship", "bleu.toml"),
		"eza.yml":       filepath.Join(themeSource, "eza", "theme.yml"),
		"bat.tmTheme":   filepath.Join(themeSource, "bat", "bleu.tmTheme"),
		"btop.theme":    filepath.Join(themeSource, "btop", "bleu.theme"),
		"yazi.toml":     filepath.Join(themeSource, "yazi", "bleu.toml"),
		"fzf.sh":        filepath.Join(themeSource, "fzf", "bleu.sh"),
	}

	for linkName, target := range symlinks {
		linkPath := filepath.Join(centralDir, linkName)

		// Use ln -snf: -s for symbolic, -n to handle existing symlink as file, -f to force overwrite
		symlinkCmd := fmt.Sprintf("su - %s -c 'ln -snf %s %s'",
			username, target, linkPath)

		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, symlinkCmd); err != nil {
			return fmt.Errorf("failed to create symlink %s -> %s: %w", linkName, target, err)
		}
	}

	return nil
}

// applyStarshipTheme configures starship to use central symlink
// Note: Actual configuration happens via STARSHIP_CONFIG env var in configs/shell/init
func (c *Configurator) applyStarshipTheme(username, relativeHome string) error {
	// Starship uses STARSHIP_CONFIG environment variable pointing to themes/current/starship.toml
	// No additional setup needed - the symlink is already created by createCentralSymlinks()
	// Shell config will be updated separately to export STARSHIP_CONFIG
	return nil
}

// applyEzaTheme creates symlink in eza config directory pointing to central theme
func (c *Configurator) applyEzaTheme(username, relativeHome string) error {
	ezaConfigDir := filepath.Join(relativeHome, ".config", "eza")
	centralTheme := filepath.Join(relativeHome, ".local", "share", "archup", "themes", "current", "eza.yml")
	ezaThemeLink := filepath.Join(ezaConfigDir, "theme.yml")

	commands := []string{
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, ezaConfigDir),
		fmt.Sprintf("su - %s -c 'ln -snf %s %s'", username, centralTheme, ezaThemeLink),
	}

	for _, cmd := range commands {
		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, cmd); err != nil {
			return fmt.Errorf("eza theme setup failed: %w", err)
		}
	}

	return nil
}

// applyBatTheme creates symlink in bat themes directory and configures bat.conf
func (c *Configurator) applyBatTheme(username, relativeHome string) error {
	batThemesDir := filepath.Join(relativeHome, ".config", "bat", "themes")
	centralTheme := filepath.Join(relativeHome, ".local", "share", "archup", "themes", "current", "bat.tmTheme")
	batThemeLink := filepath.Join(batThemesDir, "current.tmTheme")

	commands := []string{
		// Create bat themes directory
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, batThemesDir),

		// Create symlink: ~/.config/bat/themes/current.tmTheme -> themes/current/bat.tmTheme
		fmt.Sprintf("su - %s -c 'ln -snf %s %s'", username, centralTheme, batThemeLink),

		// Create bat config referencing "current" theme
		fmt.Sprintf("su - %s -c 'echo \"--theme=\\\"current\\\"\" > ~/.config/bat/bat.conf'", username),

		// Build bat cache to register the theme
		fmt.Sprintf("su - %s -c 'bat cache --build'", username),
	}

	for _, cmd := range commands {
		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, cmd); err != nil {
			return fmt.Errorf("bat theme setup failed: %w", err)
		}
	}

	return nil
}

// applyBtopTheme creates symlink in btop themes directory
func (c *Configurator) applyBtopTheme(username, relativeHome string) error {
	btopThemesDir := filepath.Join(relativeHome, ".config", "btop", "themes")
	centralTheme := filepath.Join(relativeHome, ".local", "share", "archup", "themes", "current", "btop.theme")
	btopThemeLink := filepath.Join(btopThemesDir, "current.theme")

	commands := []string{
		// Create btop themes directory
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, btopThemesDir),

		// Create symlink: ~/.config/btop/themes/current.theme -> themes/current/btop.theme
		fmt.Sprintf("su - %s -c 'ln -snf %s %s'", username, centralTheme, btopThemeLink),
	}

	for _, cmd := range commands {
		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, cmd); err != nil {
			return fmt.Errorf("btop theme setup failed: %w", err)
		}
	}

	return nil
}

// applyYaziTheme creates yazi flavor directory with symlink to central theme
func (c *Configurator) applyYaziTheme(username, relativeHome string) error {
	yaziFlavorDir := filepath.Join(relativeHome, ".config", "yazi", "flavors", "current")
	centralTheme := filepath.Join(relativeHome, ".local", "share", "archup", "themes", "current", "yazi.toml")
	yaziFlavorLink := filepath.Join(yaziFlavorDir, "flavor.toml")

	commands := []string{
		// Create yazi flavor directory for "current" theme
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, yaziFlavorDir),

		// Create symlink: ~/.config/yazi/flavors/current/flavor.toml -> themes/current/yazi.toml
		fmt.Sprintf("su - %s -c 'ln -snf %s %s'", username, centralTheme, yaziFlavorLink),
	}

	for _, cmd := range commands {
		if err := c.chrExec.ChrootExec(c.logger.LogPath(), config.PathMnt, cmd); err != nil {
			return fmt.Errorf("yazi theme setup failed: %w", err)
		}
	}

	return nil
}
