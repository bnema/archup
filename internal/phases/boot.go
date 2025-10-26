package phases

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
)

// BootPhase handles bootloader installation
type BootPhase struct {
	*BasePhase
}

// NewBootPhase creates a new boot phase
func NewBootPhase(cfg *config.Config, log *logger.Logger) *BootPhase {
	return &BootPhase{
		BasePhase: NewBasePhase("boot", "Bootloader Installation", cfg, log),
	}
}

// PreCheck validates boot prerequisites
func (p *BootPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := system.RunSimple("mountpoint", "-q", config.PathMnt)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s is not mounted", config.PathMnt)
	}

	// Verify boot partition is mounted
	result = system.RunSimple("mountpoint", "-q", config.PathMntBoot)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s is not mounted", config.PathMntBoot)
	}

	return nil
}

// Execute runs the boot phase
func (p *BootPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting bootloader installation...", 1, 4)

	// Step 1: Configure initramfs
	switch err := p.configureMkinitcpio(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Installing Limine bootloader...", 2, 4)

	// Step 2: Install Limine
	switch err := p.installLimine(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Configuring Limine...", 3, 4)

	// Step 3: Configure Limine
	switch err := p.configureLimine(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Creating UEFI boot entry...", 4, 4)

	// Step 4: Create UEFI boot entry
	switch err := p.createUEFIEntry(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Bootloader installation complete", 4, 4)
	p.SendComplete(progressChan, "Limine installed successfully")

	return PhaseResult{Success: true, Message: "Boot configuration complete"}
}

// configureMkinitcpio configures initramfs hooks
func (p *BootPhase) configureMkinitcpio(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring initramfs...")

	// Read current mkinitcpio.conf
	content, err := os.ReadFile(config.PathMntEtcMkinitcpio)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read mkinitcpio.conf: %w", err)
	}

	// Determine hooks based on encryption
	var hooks string
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		hooks = config.MkinitcpioHooksEncrypted
		p.SendOutput(progressChan, "Configuring initramfs for encrypted Plymouth boot")
	default:
		hooks = config.MkinitcpioHooksPlymouth
		p.SendOutput(progressChan, "Configuring initramfs for Plymouth boot")
	}

	// Replace HOOKS line
	hooksRegex := regexp.MustCompile(`(?m)^HOOKS=.*$`)
	newContent := hooksRegex.ReplaceAllString(string(content), hooks)

	// Write updated config
	switch err := os.WriteFile(config.PathMntEtcMkinitcpio, []byte(newContent), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write mkinitcpio.conf: %w", err)
	}

	// Regenerate initramfs
	p.SendOutput(progressChan, "Regenerating initramfs...")
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, "mkinitcpio -P"); {
	case err != nil:
		return fmt.Errorf("failed to regenerate initramfs: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Initramfs configured")

	return nil
}

// installLimine installs Limine bootloader files
func (p *BootPhase) installLimine(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Installing Limine bootloader...")

	// Install Limine to disk (BIOS - optional, may fail on UEFI-only)
	biosCmd := fmt.Sprintf("limine bios-install %s", p.config.TargetDisk)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, biosCmd); {
	case err != nil:
		p.SendOutput(progressChan, "[WARN] BIOS installation skipped (UEFI-only system)")
	default:
		p.SendOutput(progressChan, "[OK] BIOS bootloader installed")
	}

	// Create Limine directory
	switch err := os.MkdirAll(config.PathMntBootEFILimine, 0755); {
	case err != nil:
		return fmt.Errorf("failed to create Limine directory: %w", err)
	}

	// Copy Limine EFI bootloader
	srcBootloader := filepath.Join(config.PathUsrShareLimine, config.FileLimineBootloader)
	dstBootloader := filepath.Join(config.PathMntBootEFILimine, config.FileLimineBootloader)

	cpCmd := fmt.Sprintf("cp %s %s", srcBootloader, dstBootloader)
	result := p.logger.ExecCommand("sh", "-c", cpCmd)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to copy Limine bootloader: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Limine files installed")

	return nil
}

// configureLimine creates Limine configuration from template
func (p *BootPhase) configureLimine(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring Limine...")

	// Get root partition UUID
	rootPartition := p.config.RootPartition
	result := system.RunSimple("blkid", "-s", "UUID", "-o", "value", rootPartition)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to get root partition UUID: %w", result.Error)
	}

	rootUUID := strings.TrimSpace(result.Output)

	// Build kernel parameters
	var kernelParams string
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		// For encrypted setups
		kernelParams = fmt.Sprintf("cryptdevice=UUID=%s:cryptroot root=/dev/mapper/cryptroot rootflags=subvol=@ rw", rootUUID)
		p.SendOutput(progressChan, "Configured for encrypted root")
	default:
		// For non-encrypted setups
		kernelParams = fmt.Sprintf("root=UUID=%s rootflags=subvol=@ rw", rootUUID)
	}

	// Add AMD-specific kernel parameters if configured
	switch {
	case p.config.AMDPState != "":
		amdParams := fmt.Sprintf("amd_pstate=%s", p.config.AMDPState)
		kernelParams = fmt.Sprintf("%s %s", kernelParams, amdParams)
		p.SendOutput(progressChan, fmt.Sprintf("Added AMD P-State: %s", p.config.AMDPState))
	}

	// Add quiet/splash parameters
	kernelParams = fmt.Sprintf("%s %s", kernelParams, config.KernelParamsQuiet)

	// Read Limine config template
	templatePath := p.getInstallPath(config.LimineConfigTemplate)
	templateContent, err := os.ReadFile(templatePath)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read Limine template: %w", err)
	}

	// Replace template variables
	limineConfig := string(templateContent)
	limineConfig = strings.ReplaceAll(limineConfig, "{{TIMEOUT}}", config.LimineTimeout)
	limineConfig = strings.ReplaceAll(limineConfig, "{{DEFAULT_ENTRY}}", config.LimineEntry)
	limineConfig = strings.ReplaceAll(limineConfig, "{{BRANDING}}", config.LimineBranding)
	limineConfig = strings.ReplaceAll(limineConfig, "{{COLOR}}", config.LimineColor)
	limineConfig = strings.ReplaceAll(limineConfig, "{{KERNEL}}", p.config.KernelChoice)
	limineConfig = strings.ReplaceAll(limineConfig, "{{KERNEL_PARAMS}}", kernelParams)

	// Write Limine configuration
	limineConfigPath := filepath.Join(config.PathMntBootEFILimine, config.FileLimineConfig)
	switch err := os.WriteFile(limineConfigPath, []byte(limineConfig), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write Limine config: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Limine configured")

	return nil
}

// createUEFIEntry creates UEFI boot entry
func (p *BootPhase) createUEFIEntry(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Creating UEFI boot entry...")

	// Extract EFI partition number
	efiPartNum := p.extractPartitionNumber(p.config.EFIPartition)
	switch {
	case efiPartNum == "":
		return fmt.Errorf("failed to extract EFI partition number from %s", p.config.EFIPartition)
	}

	// Create UEFI boot entry
	efiCmd := fmt.Sprintf("efibootmgr --create --disk %s --part %s --label \"%s\" --loader \"%s\" --unicode",
		p.config.TargetDisk,
		efiPartNum,
		config.UEFIBootLabel,
		config.UEFIBootLoader,
	)

	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, efiCmd); {
	case err != nil:
		return fmt.Errorf("failed to create UEFI boot entry: %w", err)
	}

	p.SendOutput(progressChan, "[OK] UEFI boot entry created")

	return nil
}

// getInstallPath constructs full path to install file
func (p *BootPhase) getInstallPath(filename string) string {
	installPath := os.Getenv("ARCHUP_INSTALL")
	switch {
	case installPath == "":
		home := os.Getenv("HOME")
		installPath = filepath.Join(home, config.DefaultInstallPath)
	}
	return filepath.Join(installPath, filename)
}

// extractPartitionNumber extracts partition number from device path
func (p *BootPhase) extractPartitionNumber(partPath string) string {
	// Match partition number at end (e.g., /dev/sda1 -> 1, /dev/nvme0n1p1 -> 1)
	re := regexp.MustCompile(`[0-9]+$`)
	match := re.FindString(partPath)

	// Remove leading 'p' if present (nvme style)
	return strings.TrimPrefix(match, "p")
}

// PostCheck validates boot configuration
func (p *BootPhase) PostCheck() error {
	// Verify Limine bootloader exists
	bootloaderPath := filepath.Join(config.PathMntBootEFILimine, config.FileLimineBootloader)
	switch _, err := os.Stat(bootloaderPath); {
	case os.IsNotExist(err):
		return fmt.Errorf("Limine bootloader was not installed")
	}

	// Verify Limine config exists
	configPath := filepath.Join(config.PathMntBootEFILimine, config.FileLimineConfig)
	switch _, err := os.Stat(configPath); {
	case os.IsNotExist(err):
		return fmt.Errorf("Limine config was not created")
	}

	return p.config.Save()
}

// Rollback for boot phase
func (p *BootPhase) Rollback() error {
	// Remove Limine files if they exist
	switch err := os.RemoveAll(config.PathMntBootEFILimine); {
	case err != nil:
		p.SendOutput(nil, fmt.Sprintf("[WARN] Failed to remove Limine directory: %v", err))
	}

	return nil
}

// CanSkip returns false - boot cannot be skipped
func (p *BootPhase) CanSkip() bool {
	return false
}
