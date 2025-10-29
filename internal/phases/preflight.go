package phases

import (
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// PreflightPhase handles system validation and initial configuration
type PreflightPhase struct {
	*BasePhase
	fs      interfaces.FileSystem
	cmdExec interfaces.CommandExecutor
}

// NewPreflightPhase creates a new preflight phase
func NewPreflightPhase(cfg *config.Config, log *logger.Logger, fs interfaces.FileSystem, cmdExec interfaces.CommandExecutor) *PreflightPhase {
	return &PreflightPhase{
		BasePhase: NewBasePhase("preflight", "Preflight Checks", cfg, log),
		fs:        fs,
		cmdExec:   cmdExec,
	}
}

// PreCheck validates system requirements
func (p *PreflightPhase) PreCheck() error {
	// Check for Arch Linux
	if _, err := p.fs.Stat("/etc/arch-release"); p.fs.IsNotExist(err) {
		return fmt.Errorf("must be running on Arch Linux or Arch ISO")
	}

	// Check for derivatives
	derivatives := map[string]string{
		"/etc/cachyos-release": "CachyOS",
		"/etc/eos-release":     "EndeavourOS",
		"/etc/garuda-release":  "Garuda",
		"/etc/manjaro-release": "Manjaro",
	}

	for marker, name := range derivatives {
		if _, err := p.fs.Stat(marker); err == nil {
			return fmt.Errorf("must be vanilla Arch (detected %s)", name)
		}
	}

	// Check architecture
	out, err := p.cmdExec.Execute("uname", "-m")
	arch := strings.TrimSpace(string(out))

	switch {
	case err != nil:
		return fmt.Errorf("failed to detect architecture: %w", err)
	case arch != "x86_64":
		return fmt.Errorf("must be x86_64 architecture (detected: %s)", arch)
	}

	// Check UEFI
	if _, err := p.fs.Stat("/sys/firmware/efi/efivars"); p.fs.IsNotExist(err) {
		return fmt.Errorf("must be UEFI boot mode (legacy BIOS not supported)")
	}

	// Check Secure Boot
	out, err = p.cmdExec.Execute("bootctl", "status")
	switch {
	case err == nil && strings.Contains(string(out), "Secure Boot: enabled"):
		return fmt.Errorf("Secure Boot must be disabled")
	}

	return nil
}

// Execute runs the preflight phase
func (p *PreflightPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Running preflight checks...", 1, 2)

	// Set defaults
	p.SetDefaults()

	p.SendProgress(progressChan, "Preflight checks complete", 2, 2)
	p.SendComplete(progressChan, "Preflight complete")

	return PhaseResult{
		Success: true,
		Message: "Preflight checks passed",
	}
}

// SetDefaults sets default configuration values
func (p *PreflightPhase) SetDefaults() {
	if p.config.Hostname == "" {
		p.config.Hostname = "archup"
	}

	if p.config.Timezone == "" {
		p.config.Timezone = "UTC"
	}

	// Hardcoded bootloader
	p.config.Bootloader = config.BootloaderLimine
}

// PostCheck validates configuration after form submission
func (p *PreflightPhase) PostCheck() error {
	switch {
	case p.config.Username == "":
		return fmt.Errorf("username is required")
	case p.config.Hostname == "":
		return fmt.Errorf("hostname is required")
	case p.config.Timezone == "":
		return fmt.Errorf("timezone is required")
	}

	return p.config.Save()
}

// Rollback for preflight (no-op as no system changes made)
func (p *PreflightPhase) Rollback() error {
	return nil
}

// CanSkip returns false - preflight cannot be skipped
func (p *PreflightPhase) CanSkip() bool {
	return false
}
