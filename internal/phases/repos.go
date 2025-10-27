package phases

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// Pacman configuration constants
const (
	pacmanMultilibSectionDisabled = "#[multilib]"
	pacmanMultilibSectionEnabled  = "[multilib]"
	pacmanMultilibIncludeDisabled = "#Include = /etc/pacman.d/mirrorlist"
	pacmanMultilibIncludeEnabled  = "Include = /etc/pacman.d/mirrorlist"
	pacmanSectionPrefix           = "["
)

// AUR helper constants and validation
const (
	AURHelperParu = "paru"
	AURHelperYay  = "yay"
)

// ValidAURHelpers contains the list of supported AUR helpers
var ValidAURHelpers = []string{AURHelperParu, AURHelperYay}

// IsValidAURHelper validates if the given AUR helper is supported
func IsValidAURHelper(helper string) bool {
	if helper == "" {
		return true // Empty is valid (no AUR helper)
	}
	for _, valid := range ValidAURHelpers {
		if helper == valid {
			return true
		}
	}
	return false
}

// ReposPhase handles repository configuration
type ReposPhase struct {
	*BasePhase
	fs            interfaces.FileSystem
	sysExec       interfaces.SystemExecutor
	chrExec       interfaces.ChrootExecutor
	chaoticConfig map[string]string
}

// NewReposPhase creates a new repos phase
func NewReposPhase(cfg *config.Config, log *logger.Logger, fs interfaces.FileSystem, sysExec interfaces.SystemExecutor, chrExec interfaces.ChrootExecutor) *ReposPhase {
	return &ReposPhase{
		BasePhase:     NewBasePhase("repos", "Repository Configuration", cfg, log),
		fs:            fs,
		sysExec:       sysExec,
		chrExec:       chrExec,
		chaoticConfig: make(map[string]string),
	}
}

// PreCheck validates repos prerequisites
func (p *ReposPhase) PreCheck() error {
	// Validate AUR helper first (config validation before system checks)
	switch {
	case p.config.AURHelper != "" && !IsValidAURHelper(p.config.AURHelper):
		return fmt.Errorf("invalid AUR helper: %s (must be one of: %v)", p.config.AURHelper, ValidAURHelpers)
	}

	// Verify /mnt is mounted
	result := p.sysExec.RunSimple("mountpoint", "-q", config.PathMnt)
	switch {
	case result.ExitCode != 0:
		return fmt.Errorf("%s is not mounted", config.PathMnt)
	}

	// Verify pacman.conf exists
	switch _, err := p.fs.Stat(config.PathMntEtcPacmanConf); {
	case p.fs.IsNotExist(err):
		return fmt.Errorf("pacman.conf not found")
	}

	return nil
}

// Execute runs the repos phase
func (p *ReposPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendProgress(progressChan, "Starting repository configuration...", 1, 5)

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

	// Step 2: Sync package databases after enabling multilib
	p.SendProgress(progressChan, "Syncing package databases...", 2, 5)
	switch err := p.syncDatabases(progressChan); {
	case err != nil:
		p.logger.Warn("Database sync failed", "error", err)
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Database sync failed: %v", err))
		// Don't fail the phase, just warn
	}

	p.SendProgress(progressChan, "Installing extra packages...", 3, 5)

	// Step 3: Install extra packages (always, from official repos)
	switch err := p.installExtraPackages(progressChan); {
	case err != nil:
		// Don't fail the phase, just warn
		p.logger.Warn("Extra packages installation failed", "error", err)
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Some extra packages failed: %v", err))
	}

	p.SendProgress(progressChan, "Configuring Chaotic-AUR...", 4, 5)

	// Step 4: Enable Chaotic-AUR (optional)
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

	p.SendProgress(progressChan, "Installing AUR helper...", 5, 5)

	// Step 5: Install AUR helper (optional)
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

	p.SendProgress(progressChan, "Repository configuration complete", 5, 5)
	p.SendComplete(progressChan, "Repositories configured successfully")

	return PhaseResult{Success: true, Message: "Repos configuration complete"}
}

// syncDatabases syncs pacman package databases
func (p *ReposPhase) syncDatabases(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Syncing package databases...")

	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, "pacman -Sy --noconfirm"); {
	case err != nil:
		return fmt.Errorf("failed to sync databases: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Package databases synced")
	return nil
}

// enableMultilib enables multilib repository
func (p *ReposPhase) enableMultilib(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Enabling multilib repository...")

	// Read pacman.conf
	content, err := p.fs.ReadFile(config.PathMntEtcPacmanConf)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)

	// Check if already fully enabled (both section and Include line uncommented)
	hasMultilibSection := strings.Contains(contentStr, "[multilib]") && !strings.Contains(contentStr, "#[multilib]")
	hasMultilibInclude := strings.Contains(contentStr, pacmanMultilibIncludeEnabled)

	switch {
	case hasMultilibSection && hasMultilibInclude:
		p.SendOutput(progressChan, "[OK] Multilib already enabled")
		return nil
	}

	// Uncomment multilib section and its Include line in context-aware manner
	// This ensures we only uncomment the Include that's right after [multilib]
	lines := strings.Split(contentStr, "\n")
	inMultilib := false
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case trimmedLine == pacmanMultilibSectionDisabled:
			lines[i] = pacmanMultilibSectionEnabled
			inMultilib = true
		case trimmedLine == pacmanMultilibSectionEnabled:
			// Already enabled, track that we're in multilib section
			inMultilib = true
		case inMultilib && trimmedLine == pacmanMultilibIncludeDisabled:
			// Uncomment the Include line, preserving indentation
			lines[i] = strings.Replace(line, "#Include", "Include", 1)
			inMultilib = false // Only uncomment the first Include after [multilib]
		case inMultilib && strings.HasPrefix(trimmedLine, pacmanSectionPrefix) && trimmedLine != pacmanMultilibSectionEnabled:
			// Exited multilib section without finding Include
			inMultilib = false
		}
	}
	contentStr = strings.Join(lines, "\n")

	// Write updated config
	switch err := p.fs.WriteFile(config.PathMntEtcPacmanConf, []byte(contentStr), 0644); {
	case err != nil:
		return fmt.Errorf("failed to write pacman.conf: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Multilib enabled")

	return nil
}

// loadChaoticConfig loads Chaotic-AUR configuration from file
func (p *ReposPhase) loadChaoticConfig() error {
	configPath := p.getInstallPath(config.ChaoticConfigFile)

	file, err := p.fs.Open(configPath)
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
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, recvKeyCmd); {
	case err != nil:
		p.logger.Warn("Failed to receive Chaotic GPG key, skipping Chaotic-AUR", "error", err)
		p.SendOutput(progressChan, "[WARN] Chaotic-AUR unavailable (GPG key fetch failed)")
		return nil // Don't fail, just skip
	}

	// Locally sign the key
	lsignKeyCmd := fmt.Sprintf("pacman-key --lsign-key %s", keyID)
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, lsignKeyCmd); {
	case err != nil:
		p.logger.Warn("Failed to sign Chaotic GPG key, skipping Chaotic-AUR", "error", err)
		p.SendOutput(progressChan, "[WARN] Chaotic-AUR unavailable (GPG key signing failed)")
		return nil // Don't fail, just skip
	}

	p.SendOutput(progressChan, "[OK] Chaotic GPG key imported")

	// Download and install chaotic-keyring and chaotic-mirrorlist
	p.SendOutput(progressChan, "Downloading Chaotic-AUR packages...")

	// Parse mirror URLs (pipe-separated for fallbacks)
	var urls []string
	for _, url := range strings.Split(keyringURL, "|") {
		urls = append(urls, strings.TrimSpace(url))
	}
	for _, url := range strings.Split(mirrorlistURL, "|") {
		urls = append(urls, strings.TrimSpace(url))
	}

	switch err := p.chrExec.DownloadAndInstallPackages(p.logger.LogPath(), config.PathMnt, urls...); {
	case err != nil:
		p.logger.Warn("Chaotic-AUR packages unavailable, skipping", "error", err)
		p.SendOutput(progressChan, "[WARN] Chaotic-AUR unavailable, skipping (AUR helper will be built from source)")
		return nil // IMPORTANT: Exit early, don't add to pacman.conf
	}

	p.SendOutput(progressChan, "[OK] Chaotic packages installed")

	// Add Chaotic-AUR to pacman.conf
	content, err := p.fs.ReadFile(config.PathMntEtcPacmanConf)
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

		switch err := p.fs.WriteFile(config.PathMntEtcPacmanConf, []byte(contentStr), 0644); {
		case err != nil:
			return fmt.Errorf("failed to write pacman.conf: %w", err)
		}

		p.SendOutput(progressChan, "[OK] Added Chaotic-AUR to pacman.conf")
	}

	// Update package databases
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(),config.PathMnt, "pacman -Sy --noconfirm"); {
	case err != nil:
		return fmt.Errorf("failed to update package databases: %w", err)
	}

	// Verify Chaotic-AUR is working
	verifyCmd := fmt.Sprintf("pacman -Sl %s", repoName)
	result := p.sysExec.RunSimple("arch-chroot", config.PathMnt, "sh", "-c", verifyCmd)
	switch {
	case result.ExitCode != 0:
		p.SendOutput(progressChan, "[WARN] Chaotic-AUR verification failed")
		return fmt.Errorf("Chaotic-AUR verification failed")
	}

	p.SendOutput(progressChan, "[OK] Chaotic-AUR enabled successfully")

	return nil
}

// installPackagesIndividually tries to install packages one by one for resilience
func (p *ReposPhase) installPackagesIndividually(packages []string, progressChan chan<- ProgressUpdate) (int, []string) {
	failedPkgs := []string{}
	successCount := 0

	for _, pkg := range packages {
		individualCmd := fmt.Sprintf("pacman -S --noconfirm --needed %s", pkg)
		switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, individualCmd); {
		case err != nil:
			p.logger.Warn("Failed to install package", "package", pkg, "error", err)
			failedPkgs = append(failedPkgs, pkg)
		default:
			successCount++
		}
	}

	return successCount, failedPkgs
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

	// Try batch installation first
	installCmd := fmt.Sprintf("pacman -S --noconfirm --needed %s", strings.Join(packages, " "))
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, installCmd); {
	case err == nil:
		p.logger.Info("Extra packages installed successfully", "count", len(packages))
		p.SendOutput(progressChan, fmt.Sprintf("[OK] Installed %d extra packages", len(packages)))
		return nil
	}

	// Batch installation failed, try individual packages for resilience
	p.logger.Warn("Batch installation failed, trying individual packages", "error", err)
	p.SendOutput(progressChan, "[WARN] Batch installation failed, trying individual packages...")

	successCount, failedPkgs := p.installPackagesIndividually(packages, progressChan)

	// Report results
	switch {
	case len(failedPkgs) > 0:
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Failed to install: %s", strings.Join(failedPkgs, ", ")))
		p.logger.Warn("Some packages failed to install", "failed_count", len(failedPkgs), "packages", failedPkgs)
	}

	switch {
	case successCount > 0:
		p.SendOutput(progressChan, fmt.Sprintf("[OK] Installed %d/%d packages", successCount, len(packages)))
		p.logger.Info("Partial package installation successful", "success_count", successCount, "total", len(packages))
	default:
		p.SendOutput(progressChan, "[ERROR] Failed to install any packages")
		return fmt.Errorf("failed to install all %d packages", len(packages))
	}

	return nil
}

// loadExtraPackages reads extra package list from file
func (p *ReposPhase) loadExtraPackages() ([]string, error) {
	packageFile := p.getInstallPath(config.ExtraPackagesFile)

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

// installAURHelper installs the selected AUR helper
func (p *ReposPhase) installAURHelper(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, fmt.Sprintf("Building %s from AUR...", p.config.AURHelper))

	// Ensure build dependencies are installed
	p.SendOutput(progressChan, "Verifying build dependencies...")
	depsCmd := "pacman -S --needed --noconfirm base-devel git"
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, depsCmd); {
	case err != nil:
		p.logger.Warn("Failed to install build dependencies", "error", err)
		// Don't fail, they might already be installed
	}

	// Build AUR helper from source as configured user, then install as root
	buildCmd := fmt.Sprintf(`
		set -e

		# Build as user (clone, build, etc.)
		su - %s -c 'cd /tmp && rm -rf %s && git clone https://aur.archlinux.org/%s.git && cd %s && makepkg -s --noconfirm' || exit 1

		# Install built package as root
		pacman -U --noconfirm /tmp/%s/*.pkg.tar.zst || exit 1

		# Clean up build directory
		rm -rf /tmp/%s
	`, p.config.Username, p.config.AURHelper, p.config.AURHelper, p.config.AURHelper, p.config.AURHelper, p.config.AURHelper)

	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, buildCmd); {
	case err != nil:
		return fmt.Errorf("failed to build %s: %w", p.config.AURHelper, err)
	}

	p.SendOutput(progressChan, fmt.Sprintf("[OK] %s built and installed successfully", p.config.AURHelper))

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
	content, err := p.fs.ReadFile(config.PathMntEtcPacmanConf)
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
