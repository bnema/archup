package phases

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
)

// BaseInstallPhase handles base system installation
type BaseInstallPhase struct {
	*BasePhase
	packages []string
}

// NewBaseInstallPhase creates a new base installation phase
func NewBaseInstallPhase(cfg *config.Config, log *logger.Logger) *BaseInstallPhase {
	return &BaseInstallPhase{
		BasePhase: NewBasePhase("base", "Base System Installation", cfg, log),
	}
}

// PreCheck validates prerequisites
func (p *BaseInstallPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := system.RunSimple("mountpoint", "-q", "/mnt")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("/mnt is not mounted")
	}

	return nil
}

// Execute runs the base phase
func (p *BaseInstallPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting base installation...", 1, 5)

	// Step 1: Configure ISO pacman
	switch err := p.configurePacman(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Detecting CPU and microcode...", 2, 5)

	// Step 2: Detect CPU and select microcode
	switch err := p.detectCPU(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Installing base system...", 3, 5)

	// Step 3: Run pacstrap
	switch err := p.runPacstrap(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Generating fstab...", 4, 5)

	// Step 4: Generate fstab
	switch err := p.generateFstab(progressChan); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendProgress(progressChan, "Base installation complete", 5, 5)
	p.SendComplete(progressChan, "Base system installed successfully")

	return PhaseResult{Success: true, Message: "Base installation complete"}
}

// configurePacman enables parallel downloads on ISO
func (p *BaseInstallPhase) configurePacman(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring pacman for faster downloads...")

	// Enable ParallelDownloads=10
	result := p.logger.ExecCommand("sed", "-i",
		"s/^#ParallelDownloads = 5$/ParallelDownloads = 10/",
		"/etc/pacman.conf")

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("pacman configuration failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] Pacman configured: ParallelDownloads=10")
	return nil
}

// detectCPU detects CPU vendor and selects microcode
func (p *BaseInstallPhase) detectCPU(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Detecting CPU vendor...")

	// Read /proc/cpuinfo to detect vendor
	file, err := os.Open("/proc/cpuinfo")
	switch {
	case err != nil:
		return fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var vendor string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "vendor_id") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				vendor = parts[2]
				break
			}
		}
	}

	// Determine microcode package
	var microcode, cpuType string

	switch vendor {
	case "GenuineIntel":
		microcode = "intel-ucode"
		cpuType = "Intel"
	case "AuthenticAMD":
		microcode = "amd-ucode"
		cpuType = "AMD"
	default:
		microcode = ""
		cpuType = "Unknown"
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Unknown CPU vendor: %s", vendor))
	}

	p.config.Microcode = microcode
	p.config.CPUVendor = cpuType

	switch {
	case microcode != "":
		p.SendOutput(progressChan, fmt.Sprintf("[OK] Detected %s CPU", cpuType))
		p.SendOutput(progressChan, fmt.Sprintf("  Microcode: %s", microcode))
	}

	return nil
}

// runPacstrap installs base system with pacstrap
func (p *BaseInstallPhase) runPacstrap(progressChan chan<- ProgressUpdate) error {
	// Load base packages from embedded list
	packages, err := p.loadBasePackages()
	switch {
	case err != nil:
		return fmt.Errorf("failed to load package list: %w", err)
	}

	// Add kernel
	switch {
	case p.config.KernelChoice != "":
		packages = append(packages, p.config.KernelChoice)
		p.SendOutput(progressChan, fmt.Sprintf("Adding kernel: %s", p.config.KernelChoice))
	default:
		packages = append(packages, config.KernelLinux)
		p.config.KernelChoice = config.KernelLinux
		p.SendOutput(progressChan, "Adding kernel: linux (default)")
	}

	// Add microcode
	switch {
	case p.config.Microcode != "":
		packages = append(packages, p.config.Microcode)
		p.SendOutput(progressChan, fmt.Sprintf("Adding microcode: %s", p.config.Microcode))
	}

	// Add cryptsetup if LUKS enabled
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		packages = append(packages, "cryptsetup")
		p.SendOutput(progressChan, "Adding cryptsetup for LUKS encryption")
	}

	p.SendOutput(progressChan, fmt.Sprintf("Installing %d packages...", len(packages)))
	p.SendOutput(progressChan, "This may take several minutes...")

	// Run pacstrap
	args := append([]string{"/mnt"}, packages...)
	result := p.logger.ExecCommand("pacstrap", args...)

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("pacstrap failed: %w", result.Error)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] Installed %d packages", len(packages)))
	return nil
}

// loadBasePackages reads the base package list from file
func (p *BaseInstallPhase) loadBasePackages() ([]string, error) {
	// Read from install directory (downloaded during bootstrap)
	installPath := os.Getenv("ARCHUP_INSTALL")
	switch {
	case installPath == "":
		home := os.Getenv("HOME")
		installPath = home + "/" + config.DefaultInstallPath
	}

	packageFile := installPath + "/" + config.BasePackagesFile

	file, err := os.Open(packageFile)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to open %s: %w", packageFile, err)
	}
	defer file.Close()

	var packages []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "#"):
			continue
		}

		packages = append(packages, line)
	}

	switch err := scanner.Err(); {
	case err != nil:
		return nil, fmt.Errorf("error reading package file: %w", err)
	}

	return packages, nil
}

// generateFstab generates /etc/fstab using UUIDs
func (p *BaseInstallPhase) generateFstab(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Generating fstab with UUIDs...")

	result := system.RunSimple("sh", "-c", "genfstab -U /mnt >> /mnt/etc/fstab")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("genfstab failed: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] fstab generated")
	return nil
}

// PostCheck validates installation
func (p *BaseInstallPhase) PostCheck() error {
	// Check if base system exists
	result := system.RunSimple("test", "-d", "/mnt/usr")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("/mnt/usr does not exist")
	}

	// Check if fstab exists
	result = system.RunSimple("test", "-f", "/mnt/etc/fstab")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("/mnt/etc/fstab was not created")
	}

	return p.config.Save()
}

// Rollback removes installed system
func (p *BaseInstallPhase) Rollback() error {
	// Just log - actual cleanup happens in partitioning rollback
	p.SendOutput(nil, "[WARN] Base installation failed, partitions will be cleaned up")
	return nil
}

// CanSkip returns false - base cannot be skipped
func (p *BaseInstallPhase) CanSkip() bool {
	return false
}
