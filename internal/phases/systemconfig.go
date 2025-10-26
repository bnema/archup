package phases

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
)

// ConfigPhase handles system configuration
type ConfigPhase struct {
	*BasePhase
}

// NewConfigPhase creates a new config phase
func NewConfigPhase(cfg *config.Config, log *logger.Logger) *ConfigPhase {
	return &ConfigPhase{
		BasePhase: NewBasePhase("config", "System Configuration", cfg, log),
	}
}

// PreCheck validates configuration prerequisites
func (p *ConfigPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := system.RunSimple("mountpoint", "-q", config.PathMnt)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s is not mounted", config.PathMnt)
	}

	// Verify required config values
	switch {
	case p.config.Hostname == "":
		return fmt.Errorf("hostname is required")
	case p.config.Username == "":
		return fmt.Errorf("username is required")
	case p.config.UserPassword == "":
		return fmt.Errorf("user password is required")
	case p.config.Timezone == "":
		return fmt.Errorf("timezone is required")
	}

	return nil
}

// Execute runs the config phase
func (p *ConfigPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting system configuration...", 1, 4)

	// Step 1: System config (timezone, locale, hostname)
	switch err := p.configureSystem(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Configuring network...", 2, 4)

	// Step 2: Network config
	switch err := p.configureNetwork(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Creating user account...", 3, 4)

	// Step 3: User creation
	switch err := p.createUser(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Configuring zram swap...", 4, 4)

	// Step 4: Configure zram
	switch err := p.configureZram(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Configuration complete", 4, 4)
	p.SendComplete(progressChan, "System configured successfully")

	return PhaseResult{Success: true, Message: "Configuration complete"}
}

// configureSystem sets timezone, locale, hostname
func (p *ConfigPhase) configureSystem(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring system settings...")

	// Set timezone
	timezonePath := filepath.Join(config.PathUsrShareZoneinfo, p.config.Timezone)
	tzCmd := fmt.Sprintf("ln -sf %s %s", timezonePath, config.PathEtcLocaltime)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, tzCmd); {
	case err != nil:
		return fmt.Errorf("failed to set timezone: %w", err)
	}

	// Set hardware clock
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, "hwclock --systohc"); {
	case err != nil:
		return fmt.Errorf("failed to set hardware clock: %w", err)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] Timezone set to: %s", p.config.Timezone))

	// Set locale
	localeGen := p.config.Locale + " UTF-8\n"
	switch err := os.WriteFile(config.PathMntEtcLocaleGen, []byte(localeGen), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write locale.gen: %w", err)
	}

	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, "locale-gen"); {
	case err != nil:
		return fmt.Errorf("failed to generate locale: %w", err)
	}

	localeConf := fmt.Sprintf("LANG=%s\n", p.config.Locale)
	switch err := os.WriteFile(config.PathMntEtcLocaleConf, []byte(localeConf), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write locale.conf: %w", err)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] Locale set to: %s", p.config.Locale))

	// Set console keymap
	switch {
	case p.config.Keymap != "":
		vconsoleConf := fmt.Sprintf("KEYMAP=%s\n", p.config.Keymap)
		switch err := os.WriteFile(config.PathMntEtcVconsole, []byte(vconsoleConf), 0644); {
		case err != nil:
			return fmt.Errorf("failed to write vconsole.conf: %w", err)
		}
		p.SendOutput(progressChan, fmt.Sprintf("[OK] Console keymap set to: %s", p.config.Keymap))
	}

	// Set hostname
	switch err := os.WriteFile(config.PathMntEtcHostname, []byte(p.config.Hostname+"\n"), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write hostname: %w", err)
	}

	// Configure hosts file
	hostsContent := fmt.Sprintf(`127.0.0.1   localhost
::1         localhost
127.0.1.1   %s.localdomain %s
`, p.config.Hostname, p.config.Hostname)

	switch err := os.WriteFile(config.PathMntEtcHosts, []byte(hostsContent), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write hosts file: %w", err)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] Hostname set to: %s", p.config.Hostname))

	return nil
}

// configureNetwork enables NetworkManager
func (p *ConfigPhase) configureNetwork(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring network...")

	// Enable NetworkManager service
	switch err := system.ChrootSystemctl(p.logger.LogPath(),config.PathMnt, "enable", config.ServiceNetworkManager); {
	case err != nil:
		return fmt.Errorf("failed to enable NetworkManager: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Network configured")

	return nil
}

// createUser creates user account with sudo access
func (p *ConfigPhase) createUser(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Creating user account...")

	// Create user with home directory
	userAddCmd := fmt.Sprintf("useradd -m -G %s -s %s %s", config.GroupWheel, config.ShellBash, p.config.Username)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, userAddCmd); {
	case err != nil:
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Set user password using stdin (secure - not visible in process list)
	userPassInput := fmt.Sprintf("%s:%s", p.config.Username, p.config.UserPassword)
	switch err := system.ChrootExecWithStdin(p.logger.LogPath(), config.PathMnt, "chpasswd", userPassInput); {
	case err != nil:
		return fmt.Errorf("failed to set user password: %w", err)
	}

	// Enable sudo for wheel group (passwordless for convenience)
	switch err := os.WriteFile(config.PathMntEtcSudoersD, []byte(config.SudoersWheelContent), config.SudoersWheelPerms); {
	case err != nil:
		return fmt.Errorf("failed to write sudoers config: %w", err)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] User created: %s", p.config.Username))
	p.SendOutput(progressChan, "[OK] Sudo enabled for wheel group")

	return nil
}

// configureZram sets up zram compressed swap
func (p *ConfigPhase) configureZram(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring zram swap...")

	// Create zram-generator config
	zramConfigPath := filepath.Join(config.PathMntEtcSystemd, config.FileZramGenerator)
	switch err := os.WriteFile(zramConfigPath, []byte(config.ZramConfigContent), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write zram-generator.conf: %w", err)
	}

	// Create sysctl config for zram optimization
	sysctlConfigPath := filepath.Join(config.PathMntEtcSysctlD, config.FileSysctlZramParams)
	switch err := os.WriteFile(sysctlConfigPath, []byte(config.ZramSysctlContent), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write sysctl zram config: %w", err)
	}

	p.SendOutput(progressChan, "[OK] zram swap configured")

	return nil
}

// PostCheck validates configuration
func (p *ConfigPhase) PostCheck() error {
	// Verify hostname was set
	switch _, err := os.Stat(config.PathMntEtcHostname); {
	case os.IsNotExist(err):
		return fmt.Errorf("hostname file was not created")
	}

	// Verify locale was set
	switch _, err := os.Stat(config.PathMntEtcLocaleConf); {
	case os.IsNotExist(err):
		return fmt.Errorf("locale.conf was not created")
	}

	// Verify zram config exists
	zramConfigPath := filepath.Join(config.PathMntEtcSystemd, config.FileZramGenerator)
	switch _, err := os.Stat(zramConfigPath); {
	case os.IsNotExist(err):
		return fmt.Errorf("zram-generator.conf was not created")
	}

	return p.config.Save()
}

// Rollback for config phase (limited - no-op)
func (p *ConfigPhase) Rollback() error {
	// Configuration changes are in chroot, cleaned up by partitioning rollback
	return nil
}

// CanSkip returns false - config cannot be skipped
func (p *ConfigPhase) CanSkip() bool {
	return false
}
