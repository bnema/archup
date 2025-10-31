package verify

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
)

// ThemeVerification holds results of theme verification
type ThemeVerification struct {
	BleuThemeCloned bool
	ThemeFilesOK    map[string]bool // Tool name -> exists
	ConfigFilesOK   map[string]bool // Tool name -> config file exists
	BatCacheBuilt   bool
	StarshipValid   bool
	Warnings        []string
	Errors          []string
}

// NewThemeVerification creates a new theme verification result
func NewThemeVerification() *ThemeVerification {
	return &ThemeVerification{
		ThemeFilesOK:  make(map[string]bool),
		ConfigFilesOK: make(map[string]bool),
		Warnings:      []string{},
		Errors:        []string{},
	}
}

// ValidateBleuThemeClone verifies the bleu-theme repository was successfully cloned
func ValidateBleuThemeClone(fs interfaces.FileSystem, userHome string) bool {
	bleuThemeDir := filepath.Join(userHome, ".local", "share", "archup", "themes", "bleu")
	gitDir := filepath.Join(bleuThemeDir, ".git")

	// Check if .git directory exists (indicates successful clone)
	_, err := fs.Stat(gitDir)
	return err == nil
}

// ValidateThemeFiles checks if expected theme files exist after cloning
func ValidateThemeFiles(fs interfaces.FileSystem, userHome string) *ThemeVerification {
	result := NewThemeVerification()
	themeDir := filepath.Join(userHome, ".local", "share", "archup", "themes", "bleu")

	// Check clone first
	result.BleuThemeCloned = ValidateBleuThemeClone(fs, userHome)
	if !result.BleuThemeCloned {
		result.Warnings = append(result.Warnings, "bleu-theme not cloned successfully")
		return result
	}

	// Theme file paths to check
	themeFiles := map[string]string{
		"starship": filepath.Join(themeDir, "starship", "bleu.toml"),
		"eza":      filepath.Join(themeDir, "eza", "theme.yml"),
		"bat":      filepath.Join(themeDir, "bat", "bleu.tmTheme"),
		"btop":     filepath.Join(themeDir, "btop", "bleu.theme"),
		"yazi":     filepath.Join(themeDir, "yazi", "bleu.toml"),
		"fzf":      filepath.Join(themeDir, "fzf", "bleu.sh"),
	}

	for tool, path := range themeFiles {
		_, err := fs.Stat(path)
		exists := err == nil
		result.ThemeFilesOK[tool] = exists
		if !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s theme file missing: %s", tool, path))
		}
	}

	return result
}

// ValidateConfigFiles checks if configuration files were successfully copied
func ValidateConfigFiles(fs interfaces.FileSystem, userHome string) *ThemeVerification {
	result := NewThemeVerification()

	configFiles := map[string]string{
		"starship": filepath.Join(userHome, ".local", "share", "archup", "default", "starship.toml"),
		"eza":      filepath.Join(userHome, ".config", "eza", "theme.yml"),
		"bat":      filepath.Join(userHome, ".config", "bat", "themes", "bleu.tmTheme"),
		"btop":     filepath.Join(userHome, ".config", "btop", "themes", "bleu.theme"),
		"yazi":     filepath.Join(userHome, ".config", "yazi", "flavors", "bleu", "flavor.toml"),
	}

	for tool, path := range configFiles {
		_, err := fs.Stat(path)
		exists := err == nil
		result.ConfigFilesOK[tool] = exists
		if !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s config not found: %s", tool, path))
		}
	}

	return result
}

// ValidateBatCache checks if bat cache was built successfully
// This would require running `bat --list-themes` in chroot
func ValidateBatCache(chrExec interfaces.ChrootExecutor, logPath string) bool {
	cmd := "bat --list-themes | grep -i bleu"
	err := chrExec.ChrootExec(logPath, config.PathMnt, cmd)
	return err == nil
}

// ValidateStarshipConfig tests if starship can load its configuration
func ValidateStarshipConfig(chrExec interfaces.ChrootExecutor, logPath string) bool {
	cmd := "starship config"
	err := chrExec.ChrootExec(logPath, config.PathMnt, cmd)
	return err == nil
}

// GetThemeVerificationSummary returns a formatted summary of theme verification
func GetThemeVerificationSummary(result *ThemeVerification) string {
	var summary []string

	if result.BleuThemeCloned {
		summary = append(summary, "[OK] Bleu-theme repository cloned")
	} else {
		summary = append(summary, "[WARN] Bleu-theme repository not cloned")
	}

	// Theme files summary
	missingThemes := []string{}
	for tool, ok := range result.ThemeFilesOK {
		if !ok {
			missingThemes = append(missingThemes, tool)
		}
	}
	if len(missingThemes) > 0 {
		summary = append(summary, fmt.Sprintf("[WARN] Missing theme files: %s", strings.Join(missingThemes, ", ")))
	} else if len(result.ThemeFilesOK) > 0 {
		summary = append(summary, "[OK] All theme files found")
	}

	// Config files summary
	missingConfigs := []string{}
	for tool, ok := range result.ConfigFilesOK {
		if !ok {
			missingConfigs = append(missingConfigs, tool)
		}
	}
	if len(missingConfigs) > 0 {
		summary = append(summary, fmt.Sprintf("[WARN] Missing config files: %s", strings.Join(missingConfigs, ", ")))
	} else if len(result.ConfigFilesOK) > 0 {
		summary = append(summary, "[OK] All config files found")
	}

	// Bat cache
	if result.BatCacheBuilt {
		summary = append(summary, "[OK] Bat cache built successfully")
	}

	// Starship config
	if result.StarshipValid {
		summary = append(summary, "[OK] Starship config valid")
	}

	// Warnings
	for _, w := range result.Warnings {
		summary = append(summary, fmt.Sprintf("[WARN] %s", w))
	}

	// Errors
	for _, e := range result.Errors {
		summary = append(summary, fmt.Sprintf("[ERROR] %s", e))
	}

	return strings.Join(summary, "\n")
}
