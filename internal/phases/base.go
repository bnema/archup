package phases

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// BaseInstallPhase handles base system installation
type BaseInstallPhase struct {
	*BasePhase
	fs      interfaces.FileSystem
	sysExec interfaces.SystemExecutor
	packages []string
}

// NewBaseInstallPhase creates a new base installation phase
func NewBaseInstallPhase(cfg *config.Config, log *logger.Logger, fs interfaces.FileSystem, sysExec interfaces.SystemExecutor) *BaseInstallPhase {
	return &BaseInstallPhase{
		BasePhase: NewBasePhase("base", "Base System Installation", cfg, log),
		fs:        fs,
		sysExec:   sysExec,
	}
}

// PreCheck validates prerequisites
func (p *BaseInstallPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := p.sysExec.RunSimple("mountpoint", "-q", "/mnt")
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

	// Step 2.5: Configure CachyOS repository on ISO if needed
	switch {
	case p.config.KernelChoice == config.KernelLinuxCachyOS:
		p.SendProgress(progressChan, "Configuring CachyOS repository...", 2, 5)
		switch err := p.configureCachyOSRepo(progressChan); {
		case err != nil:
			p.SendError(progressChan, err)
			return PhaseResult{Success: false, Error: err}
		}
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

// configureCachyOSRepo configures CachyOS repository on the ISO before pacstrap
func (p *BaseInstallPhase) configureCachyOSRepo(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring CachyOS repository on live system...")

	// Import CachyOS GPG key
	p.SendOutput(progressChan, "Importing CachyOS GPG key...")
	result := p.logger.ExecCommand("pacman-key", "--recv-keys", "F3B607488DB35A47", "--keyserver", "keyserver.ubuntu.com")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to receive CachyOS GPG key: %w", result.Error)
	}

	result = p.logger.ExecCommand("pacman-key", "--lsign-key", "F3B607488DB35A47")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to sign CachyOS GPG key: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] GPG key imported")

	// Check if CachyOS repo already configured
	content, err := p.fs.ReadFile("/etc/pacman.conf")
	switch {
	case err != nil:
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[cachyos]") {
		// Add CachyOS repo before [core] repository
		p.SendOutput(progressChan, "Adding CachyOS repository to pacman.conf...")

		// Find [core] and insert before it
		lines := strings.Split(contentStr, "\n")
		var newLines []string
		inserted := false

		for _, line := range lines {
			if !inserted && strings.TrimSpace(line) == "[core]" {
				newLines = append(newLines, "# CachyOS repositories")
				newLines = append(newLines, "[cachyos]")
				newLines = append(newLines, "Include = /etc/pacman.d/cachyos-mirrorlist")
				newLines = append(newLines, "")
				inserted = true
			}
			newLines = append(newLines, line)
		}

		contentStr = strings.Join(newLines, "\n")

		switch err := p.fs.WriteFile("/etc/pacman.conf", []byte(contentStr), 0644); {
		case err != nil:
			return fmt.Errorf("failed to write pacman.conf: %w", err)
		}
	}

	// Create CachyOS mirrorlist
	p.SendOutput(progressChan, "Creating CachyOS mirrorlist...")
	mirrorlist := "## CachyOS mirrorlist\nServer = https://mirror.cachyos.org/repo/$arch/$repo\n"

	switch err := p.fs.WriteFile("/etc/pacman.d/cachyos-mirrorlist", []byte(mirrorlist), 0644); {
	case err != nil:
		return fmt.Errorf("failed to create cachyos-mirrorlist: %w", err)
	}

	// Sync databases
	p.SendOutput(progressChan, "Syncing package databases...")
	result = p.logger.ExecCommand("pacman", "-Sy")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to sync databases: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] CachyOS repository configured")
	return nil
}

// copyCachyOSConfig copies CachyOS repository configuration to installed system
func (p *BaseInstallPhase) copyCachyOSConfig(progressChan chan<- ProgressUpdate) error {
	// Read ISO's pacman.conf to check if CachyOS is configured
	content, err := p.fs.ReadFile("/etc/pacman.conf")
	switch {
	case err != nil:
		return fmt.Errorf("failed to read ISO pacman.conf: %w", err)
	}

	isoContent := string(content)

	// Read installed system's pacman.conf
	mntContent, err := p.fs.ReadFile("/mnt/etc/pacman.conf")
	switch {
	case err != nil:
		return fmt.Errorf("failed to read /mnt/etc/pacman.conf: %w", err)
	}

	mntContentStr := string(mntContent)

	// Add CachyOS repo before [core] if not already present
	if !strings.Contains(mntContentStr, "[cachyos]") && strings.Contains(isoContent, "[cachyos]") {
		lines := strings.Split(mntContentStr, "\n")
		var newLines []string
		inserted := false

		for _, line := range lines {
			if !inserted && strings.TrimSpace(line) == "[core]" {
				newLines = append(newLines, "# CachyOS repositories")
				newLines = append(newLines, "[cachyos]")
				newLines = append(newLines, "Include = /etc/pacman.d/cachyos-mirrorlist")
				newLines = append(newLines, "")
				inserted = true
			}
			newLines = append(newLines, line)
		}

		mntContentStr = strings.Join(newLines, "\n")

		switch err := p.fs.WriteFile("/mnt/etc/pacman.conf", []byte(mntContentStr), 0644); {
		case err != nil:
			return fmt.Errorf("failed to write /mnt/etc/pacman.conf: %w", err)
		}

		p.SendOutput(progressChan, "[OK] Added CachyOS to installed system's pacman.conf")
	}

	// Copy mirrorlist from ISO to installed system
	result := p.logger.ExecCommand("mkdir", "-p", "/mnt/etc/pacman.d")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to create /mnt/etc/pacman.d: %w", result.Error)
	}

	result = p.logger.ExecCommand("cp", "/etc/pacman.d/cachyos-mirrorlist", "/mnt/etc/pacman.d/cachyos-mirrorlist")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("failed to copy cachyos-mirrorlist: %w", result.Error)
	}

	p.SendOutput(progressChan, "[OK] CachyOS repository configured in installed system")
	return nil
}

// detectCPU detects CPU vendor and selects microcode
func (p *BaseInstallPhase) detectCPU(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Detecting CPU vendor...")

	// Read /proc/cpuinfo to detect vendor
	file, err := p.fs.Open("/proc/cpuinfo")
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

	p.SendOutput(progressChan, fmt.Sprintf("Installing %d packages:", len(packages)))

	// Show package list
	for _, pkg := range packages {
		p.SendOutput(progressChan, fmt.Sprintf("  - %s", pkg))
	}

	p.SendOutput(progressChan, "")
	p.SendOutput(progressChan, "Starting installation (this may take several minutes)...")

	// Run pacstrap
	args := append([]string{"/mnt"}, packages...)
	result := p.logger.ExecCommand("pacstrap", args...)

	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("pacstrap failed: %w", result.Error)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] Installed %d packages", len(packages)))

	// Copy CachyOS configuration to installed system if needed
	switch {
	case p.config.KernelChoice == config.KernelLinuxCachyOS:
		p.SendOutput(progressChan, "Configuring CachyOS repository in installed system...")
		if err := p.copyCachyOSConfig(progressChan); err != nil {
			return fmt.Errorf("failed to configure CachyOS in installed system: %w", err)
		}
	}

	return nil
}

// loadBasePackages reads the base package list from file
func (p *BaseInstallPhase) loadBasePackages() ([]string, error) {
	// Read from install directory (downloaded during bootstrap)
	// Use DefaultInstallDir directly to match where bootstrap downloads files
	packageFile := config.DefaultInstallDir + "/" + config.BasePackagesFile

	file, err := p.fs.Open(packageFile)
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

	result := p.sysExec.RunSimple("sh", "-c", "genfstab -U /mnt >> /mnt/etc/fstab")
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
	result := p.sysExec.RunSimple("test", "-d", "/mnt/usr")
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("/mnt/usr does not exist")
	}

	// Check if fstab exists
	result = p.sysExec.RunSimple("test", "-f", "/mnt/etc/fstab")
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
