package verify

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
)

// ShellConfig holds verification results for shell configuration
type ShellConfig struct {
	BashrcExists    bool
	ShellFilesOK    bool
	BashrcSyntaxOK  bool
	Warnings        []string
	Errors          []string
}

// ValidateShellConfigs performs comprehensive shell configuration verification
func ValidateShellConfigs(
	fs interfaces.FileSystem,
	chrExec interfaces.ChrootExecutor,
	logPath string,
	userHome string,
) ShellConfig {
	result := ShellConfig{
		Warnings: []string{},
		Errors:   []string{},
	}

	// Check .bashrc exists
	bashrcPath := filepath.Join(userHome, ".bashrc")
	_, err := fs.Stat(bashrcPath)
	if fs.IsNotExist(err) {
		result.BashrcExists = false
		result.Warnings = append(result.Warnings, ".bashrc not found")
	} else if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf(".bashrc stat error: %v", err))
	} else {
		result.BashrcExists = true
	}

	// Check shell config directory exists
	archupDefaultBash := filepath.Join(userHome, ".local", "share", "archup", "default", "bash")
	_, err = fs.Stat(archupDefaultBash)
	if fs.IsNotExist(err) {
		result.Warnings = append(result.Warnings, "shell config directory missing")
	} else if err == nil {
		result.ShellFilesOK = true
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("shell config directory stat error: %v", err))
	}

	return result
}

// ValidateBashSyntax tests bash configuration syntax by loading bash in chroot
func ValidateBashSyntax(
	chrExec interfaces.ChrootExecutor,
	logPath string,
) error {
	// Test bash syntax: bash -n ~/.bashrc in chroot
	cmd := "bash -n /root/.bashrc"
	if err := chrExec.ChrootExec(logPath, config.PathMnt, cmd); err != nil {
		return fmt.Errorf("bash syntax check failed: %w", err)
	}
	return nil
}

// GetShellConfigWarnings returns formatted warning messages
func GetShellConfigWarnings(result ShellConfig) string {
	if len(result.Warnings) == 0 && len(result.Errors) == 0 {
		return ""
	}

	var msgs []string
	for _, w := range result.Warnings {
		msgs = append(msgs, fmt.Sprintf("[WARN] %s", w))
	}
	for _, e := range result.Errors {
		msgs = append(msgs, fmt.Sprintf("[ERROR] %s", e))
	}

	return strings.Join(msgs, "\n")
}
