package phases

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/system"
)

// ReposPhase handles repository configuration
type ReposPhase struct {
	*BasePhase
	chaoticConfig map[string]string
}

// NewReposPhase creates a new repos phase
func NewReposPhase(cfg *config.Config, log *logger.Logger) *ReposPhase {
	return &ReposPhase{
		BasePhase:     NewBasePhase("repos", "Repository Configuration", cfg, log),
		chaoticConfig: make(map[string]string),
	}
}

// PreCheck validates repos prerequisites
func (p *ReposPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := system.RunSimple("mountpoint", "-q", config.PathMnt)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s is not mounted", config.PathMnt)
	}

	// Verify pacman.conf exists
	switch _, err := os.Stat(config.PathMntEtcPacmanConf); {
	case os.IsNotExist(err):
		return fmt.Errorf("pacman.conf not found")
	}

	return nil
}

// Execute runs the repos phase
func (p *ReposPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting repository configuration...", 1, 4)

	// Step 1: Enable multilib (optional)
	switch {
	case p.config.EnableMultilib:
		switch err := p.enableMultilib(progressChan); {
		case err != nil:
			p.SendError(progressChan, err)
			return PhaseResult{Success: false, Error: err}
		}
	default:
		p.SendOutput(progressChan, "[SKIP] Multilib disabled")
	}

	p.SendProgress(progressChan, "Installing extra packages...", 2, 4)

	// Step 2: Install extra packages (always, from official repos)
	switch err := p.installExtraPackages(progressChan); {
	case err != nil:
		// Don't fail the phase, just warn
		p.logger.Warn("Extra packages installation failed", "error", err)
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Some extra packages failed: %v", err))
	}

	p.SendProgress(progressChan, "Configuring Chaotic-AUR...", 3, 4)

	// Step 3: Enable Chaotic-AUR (optional)
	switch {
	case p.config.EnableChaotic:
		switch err := p.enableChaoticAUR(progressChan); {
		case err != nil:
			p.SendError(progressChan, err)
			return PhaseResult{Success: false, Error: err}
		}
	default:
		p.SendOutput(progressChan, "[SKIP] Chaotic-AUR disabled")
	}

	p.SendProgress(progressChan, "Installing AUR helper...", 4, 4)

	// Step 4: Install AUR helper (optional)
	switch {
	case p.config.AURHelper != "":
		switch err := p.installAURHelper(progressChan); {
		case err != nil:
			p.SendError(progressChan, err)
			return PhaseResult{Success: false, Error: err}
		}
	default:
		p.SendOutput(progressChan, "[SKIP] AUR helper disabled")
	}

	p.SendProgress(progressChan, "Repository configuration complete", 4, 4)
	p.SendComplete(progressChan, "Repositories configured successfully")

	return PhaseResult{Success: true, Message: "Repos configuration complete"}
}

// enableMultilib enables multilib repository
func (p *ReposPhase) enableMultilib(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Enabling multilib repository...")

	// Read pacman.conf
	content, err := os.ReadFile(config.PathMntEtcPacmanConf)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)

	// Check if already enabled
	switch {
	case strings.Contains(contentStr, "[multilib]") && !strings.Contains(contentStr, "#[multilib]"):
		p.SendOutput(progressChan, "[OK] Multilib already enabled")
		return nil
	}

	// Uncomment multilib section
	multilibRegex := regexp.MustCompile(`(?m)^#\[multilib\]`)
	contentStr = multilibRegex.ReplaceAllString(contentStr, "[multilib]")

	includeRegex := regexp.MustCompile(`(?m)^#Include = /etc/pacman.d/mirrorlist`)
	contentStr = includeRegex.ReplaceAllString(contentStr, "Include = /etc/pacman.d/mirrorlist")

	// Write updated config
	switch err := os.WriteFile(config.PathMntEtcPacmanConf, []byte(contentStr), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write pacman.conf: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Multilib enabled")

	return nil
}

// loadChaoticConfig loads Chaotic-AUR configuration from file
func (p *ReposPhase) loadChaoticConfig() error {
	configPath := p.getInstallPath(config.ChaoticConfigFile)

	file, err := os.Open(configPath)
	switch {
	case err != nil:
		return fmt.Errorf("failed to open chaotic config: %w", err)
	}
	defer file.Close()

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

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		switch {
		case len(parts) == 2:
			p.chaoticConfig[parts[0]] = parts[1]
		}
	}

	switch err := scanner.Err(); {
	case err != nil:
		return fmt.Errorf("error reading chaotic config: %w", err)
	}

	return nil
}

// enableChaoticAUR enables Chaotic-AUR repository
func (p *ReposPhase) enableChaoticAUR(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Adding Chaotic-AUR repository...")

	// Load configuration
	switch err := p.loadChaoticConfig(); {
	case err != nil:
		return fmt.Errorf("failed to load Chaotic config: %w", err)
	}

	keyID := p.chaoticConfig["CHAOTIC_KEY_ID"]
	keyserver := p.chaoticConfig["CHAOTIC_KEYSERVER"]
	keyringURL := p.chaoticConfig["CHAOTIC_KEYRING_URL"]
	mirrorlistURL := p.chaoticConfig["CHAOTIC_MIRRORLIST_URL"]
	repoName := p.chaoticConfig["CHAOTIC_REPO_NAME"]
	mirrorlistPath := p.chaoticConfig["CHAOTIC_MIRRORLIST_PATH"]

	// Import GPG key
	recvKeyCmd := fmt.Sprintf("pacman-key --recv-key %s --keyserver %s", keyID, keyserver)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, recvKeyCmd); {
	case err != nil:
		return fmt.Errorf("failed to receive Chaotic GPG key: %w", err)
	}

	// Locally sign the key
	lsignKeyCmd := fmt.Sprintf("pacman-key --lsign-key %s", keyID)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, lsignKeyCmd); {
	case err != nil:
		return fmt.Errorf("failed to sign Chaotic GPG key: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Chaotic GPG key imported")

	// Install chaotic-keyring and chaotic-mirrorlist
	installCmd := fmt.Sprintf("pacman -U --noconfirm '%s' '%s'", keyringURL, mirrorlistURL)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, installCmd); {
	case err != nil:
		return fmt.Errorf("failed to install Chaotic packages: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Chaotic packages installed")

	// Add Chaotic-AUR to pacman.conf
	content, err := os.ReadFile(config.PathMntEtcPacmanConf)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)

	// Check if already added
	switch {
	case strings.Contains(contentStr, fmt.Sprintf("[%s]", repoName)):
		p.SendOutput(progressChan, "[OK] Chaotic-AUR already in pacman.conf")
	default:
		// Append Chaotic-AUR section
		chaoticSection := fmt.Sprintf("\n# Chaotic-AUR repository\n[%s]\nInclude = %s\n", repoName, mirrorlistPath)
		contentStr += chaoticSection

		switch err := os.WriteFile(config.PathMntEtcPacmanConf, []byte(contentStr), 0644); {
		case err != nil:
			return fmt.Errorf("failed to write pacman.conf: %w", err)
		}

		p.SendOutput(progressChan, "[OK] Added Chaotic-AUR to pacman.conf")
	}

	// Update package databases
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, "pacman -Sy --noconfirm"); {
	case err != nil:
		return fmt.Errorf("failed to update package databases: %w", err)
	}

	// Verify Chaotic-AUR is working
	verifyCmd := fmt.Sprintf("pacman -Sl %s", repoName)
	result := system.RunSimple("arch-chroot", config.PathMnt, "sh", "-c", verifyCmd)
	switch {
	case result.ExitCode != 0:
		p.SendOutput(progressChan, "[WARN] Chaotic-AUR verification failed")
		return fmt.Errorf("Chaotic-AUR verification failed")
	}

	p.SendOutput(progressChan, "[OK] Chaotic-AUR enabled successfully")

	return nil
}

// installExtraPackages installs packages from extra.packages file
func (p *ReposPhase) installExtraPackages(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Installing extra packages...")

	// Load extra packages
	packages, err := p.loadExtraPackages()
	switch {
	case err != nil:
		return fmt.Errorf("failed to load extra packages: %w", err)
	case len(packages) == 0:
		p.SendOutput(progressChan, "[SKIP] No extra packages to install")
		return nil
	}

	p.logger.Info("Installing extra packages", "count", len(packages), "packages", strings.Join(packages, ", "))
	p.SendOutput(progressChan, fmt.Sprintf("Installing %d extra packages...", len(packages)))

	// Install packages
	installCmd := fmt.Sprintf("pacman -S --noconfirm %s", strings.Join(packages, " "))
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, installCmd); {
	case err != nil:
		p.logger.Error("Failed to install extra packages", "error", err, "packages", strings.Join(packages, ", "))
		return fmt.Errorf("failed to install extra packages: %w", err)
	}

	p.logger.Info("Extra packages installed successfully", "count", len(packages))
	p.SendOutput(progressChan, fmt.Sprintf("[OK] Installed %d extra packages", len(packages)))

	return nil
}

// loadExtraPackages reads extra package list from file
func (p *ReposPhase) loadExtraPackages() ([]string, error) {
	packageFile := p.getInstallPath(config.ExtraPackagesFile)

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

// installAURHelper installs the selected AUR helper
func (p *ReposPhase) installAURHelper(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, fmt.Sprintf("Installing %s AUR helper...", p.config.AURHelper))

	// Install AUR helper from Chaotic-AUR
	installCmd := fmt.Sprintf("pacman -S --noconfirm %s", p.config.AURHelper)
	switch err := system.ChrootExec(p.logger.LogPath(),config.PathMnt, installCmd); {
	case err != nil:
		return fmt.Errorf("failed to install %s: %w", p.config.AURHelper, err)
	}

	// Verify installation
	verifyCmd := fmt.Sprintf("su - %s -c '%s --version'", p.config.Username, p.config.AURHelper)
	result := system.RunSimple("arch-chroot", config.PathMnt, "sh", "-c", verifyCmd)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s verification failed", p.config.AURHelper)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] %s installed successfully", p.config.AURHelper))

	return nil
}

// getInstallPath constructs full path to install file
func (p *ReposPhase) getInstallPath(filename string) string {
	// Use DefaultInstallDir directly to match where bootstrap downloads files
	return filepath.Join(config.DefaultInstallDir, filename)
}

// PostCheck validates repository configuration
func (p *ReposPhase) PostCheck() error {
	// Verify pacman.conf was modified
	content, err := os.ReadFile(config.PathMntEtcPacmanConf)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)

	// Check multilib if enabled
	switch {
	case p.config.EnableMultilib:
		switch {
		case !strings.Contains(contentStr, "[multilib]"):
			return fmt.Errorf("multilib was not enabled in pacman.conf")
		}
	}

	// Check Chaotic-AUR if enabled
	switch {
	case p.config.EnableChaotic:
		repoName := p.chaoticConfig["CHAOTIC_REPO_NAME"]
		switch {
		case repoName == "":
			repoName = "chaotic-aur"
		}

		switch {
		case !strings.Contains(contentStr, fmt.Sprintf("[%s]", repoName)):
			return fmt.Errorf("Chaotic-AUR was not added to pacman.conf")
		}
	}

	return p.config.Save()
}

// Rollback for repos phase
func (p *ReposPhase) Rollback() error {
	// Repository changes are in chroot, cleaned up by partitioning rollback
	return nil
}

// CanSkip returns false - repos cannot be skipped
func (p *ReposPhase) CanSkip() bool {
	return false
}
