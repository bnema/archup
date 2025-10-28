package phases

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/postboot"
	"github.com/bnema/archup/internal/shell"
	"github.com/bnema/archup/internal/system"
	"github.com/bnema/archup/internal/verify"
)

// PostInstallPhase handles post-installation tasks
type PostInstallPhase struct {
	*BasePhase
	fs      interfaces.FileSystem
	http    interfaces.HTTPClient
	sysExec interfaces.SystemExecutor
	chrExec interfaces.ChrootExecutor
}

// NewPostInstallPhase creates a new post-install phase
func NewPostInstallPhase(cfg *config.Config, log *logger.Logger, fs interfaces.FileSystem, http interfaces.HTTPClient, sysExec interfaces.SystemExecutor, chrExec interfaces.ChrootExecutor) *PostInstallPhase {
	return &PostInstallPhase{
		BasePhase: NewBasePhase("postinstall", "Post-Installation", cfg, log),
		fs:        fs,
		http:      http,
		sysExec:   sysExec,
		chrExec:   chrExec,
	}
}

// PreCheck validates post-install prerequisites
func (p *PostInstallPhase) PreCheck() error {
	// Verify /mnt is mounted
	result := p.sysExec.RunSimple("mountpoint", "-q", config.PathMnt)
	if result.ExitCode != 0 {
		return fmt.Errorf("%s is not mounted", config.PathMnt)
	}

	// Verify boot directory exists
	if _, err := p.fs.Stat(config.PathMntBoot); p.fs.IsNotExist(err) {
		return fmt.Errorf("boot directory not found")
	}

	return nil
}

// Execute runs the post-install phase
func (p *PostInstallPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	totalSteps := 10
	currentStep := 0

	// Step 1: Boot logo
	currentStep++
	p.SendProgress(progressChan, "Installing boot logo...", currentStep, totalSteps)
	if err := p.installBootLogo(progressChan); err != nil {
		// Non-fatal, continue
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Boot logo failed: %v", err))
	}

	// Step 2: Plymouth
	currentStep++
	p.SendProgress(progressChan, "Configuring Plymouth splash screen...", currentStep, totalSteps)
	if err := p.configurePlymouth(progressChan); err != nil {
		// Non-fatal, continue
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Plymouth failed: %v", err))
	}

	// Step 3: Snapper
	currentStep++
	p.SendProgress(progressChan, "Configuring btrfs snapshots...", currentStep, totalSteps)
	if err := p.configureSnapper(progressChan); err != nil {
		// Non-fatal, continue
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Snapper failed: %v", err))
	}

	// Step 4: Pacman configuration
	currentStep++
	p.SendProgress(progressChan, "Configuring pacman...", currentStep, totalSteps)
	if err := p.configurePacman(progressChan); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	// Step 5: Post-boot setup
	currentStep++
	p.SendProgress(progressChan, "Setting up post-boot scripts...", currentStep, totalSteps)
	postbootSetup := postboot.NewSetup(p.fs, p.http, p.chrExec, p.config, p.logger)
	pbResult, err := postbootSetup.Configure()
	if err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}
	if pbResult.ScriptsDownloaded > 0 {
		p.SendOutput(progressChan, fmt.Sprintf("[OK] Post-boot scripts downloaded (%d files)", pbResult.ScriptsDownloaded))
	}
	if pbResult.ServiceEnabled {
		p.SendOutput(progressChan, "[OK] First-boot service enabled")
	}

	// Step 6: Pacman hooks
	currentStep++
	p.SendProgress(progressChan, "Installing pacman hooks...", currentStep, totalSteps)
	if err := p.installPacmanHooks(progressChan); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	// Step 7: Shell configuration
	currentStep++
	p.SendProgress(progressChan, "Configuring shell environment...", currentStep, totalSteps)
	shellConfig := shell.NewConfigurator(p.fs, p.http, p.chrExec, p.config, p.logger)

	// Get username from /mnt/home
	homeDir := filepath.Join(config.PathMnt, "home")
	entries, err := os.ReadDir(homeDir)
	switch {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: fmt.Errorf("failed to read home directory: %w", err)}
	case len(entries) == 0:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: fmt.Errorf("no user found in /home")}
	}

	username := entries[0].Name()
	userHome := filepath.Join(config.PathMnt, "home", username)

	shellResult, err := shellConfig.Configure(username, userHome)
	if err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}
	p.SendOutput(progressChan, fmt.Sprintf("[OK] Shell environment configured (%d themes applied)", shellResult.ThemesApplied))
	if len(shellResult.ThemesFailed) > 0 {
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Failed to apply themes: %v", shellResult.ThemesFailed))
	}
	for _, w := range shellResult.Warnings {
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] %s", w))
	}

	// Step 7a: Shell configuration verification (non-fatal)
	currentStep++
	p.SendProgress(progressChan, "Verifying shell configuration...", currentStep, totalSteps)
	shellVerification := verify.ValidateShellConfigs(p.fs, p.chrExec, p.logger.LogPath(), userHome)
	if warnings := verify.GetShellConfigWarnings(shellVerification); warnings != "" {
		p.SendOutput(progressChan, warnings)
	}

	// Step 7b: Theme verification (non-fatal)
	themeFiles := verify.ValidateThemeFiles(p.fs, userHome)
	p.SendOutput(progressChan, verify.GetThemeVerificationSummary(themeFiles))

	// Step 8: Installation verification
	currentStep++
	p.SendProgress(progressChan, "Verifying installation...", currentStep, totalSteps)
	if err := p.verifyInstallation(progressChan); err != nil {
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Verification warnings: %v", err))
	}

	// Step 9: Unmount
	currentStep++
	p.SendProgress(progressChan, "Unmounting filesystems...", currentStep, totalSteps)
	if err := p.unmountAll(progressChan); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	p.SendComplete(progressChan, "Post-installation complete")
	return PhaseResult{Success: true, Message: "Post-installation complete"}
}

// installBootLogo downloads and configures Limine boot logo
func (p *PostInstallPhase) installBootLogo(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Downloading Arch logo...")

	// Build logo URL from config
	logoURL := fmt.Sprintf("%s/%s", p.config.RawURL, config.ArchLogoURL)

	resp, err := p.http.Get(logoURL)
	if err != nil {
		return fmt.Errorf("failed to download logo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download logo: HTTP %d", resp.StatusCode)
	}

	// Save to boot partition
	logoFile, err := p.fs.Create(config.PathMntBootLogo)
	if err != nil {
		return fmt.Errorf("failed to create logo file: %w", err)
	}
	defer logoFile.Close()

	_, err = io.Copy(logoFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save logo: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Logo downloaded")

	// Update Limine config
	limineConf := filepath.Join(config.PathMntBootEFILimine, config.FileLimineConfig)
	content, err := p.fs.ReadFile(limineConf)
	if err != nil {
		return fmt.Errorf("failed to read limine.conf: %w", err)
	}

	contentStr := string(content)

	// Find "graphics: yes" and add wallpaper settings after it
	graphicsRegex := regexp.MustCompile(`(?m)^graphics: yes$`)
	switch {
	case graphicsRegex.MatchString(contentStr):
		wallpaperSettings := "\nwallpaper: boot():/arch-logo.png\nwallpaper_style: centered\nbackdrop: 000000"
		contentStr = graphicsRegex.ReplaceAllString(contentStr, "graphics: yes"+wallpaperSettings)

		if err := p.fs.WriteFile(limineConf, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to update limine.conf: %w", err)
		}

		p.SendOutput(progressChan, "[OK] Boot logo configured")
	default:
		p.SendOutput(progressChan, "[WARN] graphics: yes not found in limine.conf")
	}

	return nil
}

// configurePlymouth sets up Plymouth splash screen
func (p *PostInstallPhase) configurePlymouth(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Setting up Plymouth theme...")

	// Create theme directory
	themeDir := filepath.Join(config.PathMntPlymouthThemes, config.PlymouthThemeName)
	if err := p.fs.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	// Download all Plymouth files from assets (local files in repo)
	for _, filename := range config.PlymouthFiles {
		p.SendOutput(progressChan, fmt.Sprintf("Downloading %s...", filename))

		fileURL := fmt.Sprintf("%s/assets/plymouth/%s", p.config.RawURL, filename)
		resp, err := p.http.Get(fileURL)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", filename, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("failed to download %s: HTTP %d", filename, resp.StatusCode)
		}

		// Save file
		destPath := filepath.Join(themeDir, filename)
		destFile, err := p.fs.Create(destPath)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}

		_, err = io.Copy(destFile, resp.Body)
		destFile.Close()
		resp.Body.Close()

		if err != nil {
			return fmt.Errorf("failed to save %s: %w", filename, err)
		}

		// Set permissions
		if err := p.fs.Chmod(destPath, 0644); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", filename, err)
		}
	}

	p.SendOutput(progressChan, "[OK] Plymouth files downloaded")

	// Set as default theme
	setThemeCmd := fmt.Sprintf("plymouth-set-default-theme %s", config.PlymouthThemeName)
	if err := p.chrExec.ChrootExec(p.logger.LogPath(),config.PathMnt, setThemeCmd); err != nil {
		return fmt.Errorf("failed to set Plymouth theme: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Plymouth theme set")

	// Regenerate initramfs
	p.SendOutput(progressChan, "Regenerating initramfs with Plymouth...")
	if err := p.chrExec.ChrootExec(p.logger.LogPath(),config.PathMnt, "mkinitcpio -P"); err != nil {
		return fmt.Errorf("failed to regenerate initramfs: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Plymouth configured")
	return nil
}

// configureSnapper sets up btrfs snapshots with Limine integration
func (p *PostInstallPhase) configureSnapper(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Installing limine-snapper-sync...")

	// Install limine-snapper-sync
	if err := p.chrExec.ChrootPacman(p.logger.LogPath(),config.PathMnt, "-S", "--needed", "limine-snapper-sync"); err != nil {
		return fmt.Errorf("failed to install limine-snapper-sync: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Package installed")

	// Get kernel cmdline from existing Limine config
	limineConf := filepath.Join(config.PathMntBootEFILimine, config.FileLimineConfig)
	content, err := p.fs.ReadFile(limineConf)
	if err != nil {
		return fmt.Errorf("failed to read limine.conf: %w", err)
	}

	// Extract cmdline
	cmdlineRegex := regexp.MustCompile(`(?m)^\s*cmdline:\s*(.+)$`)
	matches := cmdlineRegex.FindStringSubmatch(string(content))
	cmdline := ""
	if len(matches) > 1 {
		cmdline = strings.TrimSpace(matches[1])
	}

	// Create Limine snapper config
	limineDefaultConfig := fmt.Sprintf(`TARGET_OS_NAME="ArchUp"

ESP_PATH="/boot"

KERNEL_CMDLINE[default]="%s"

ENABLE_UKI=no
ENABLE_LIMINE_FALLBACK=yes

FIND_BOOTLOADERS=yes

BOOT_ORDER="*, *fallback, Snapshots"

MAX_SNAPSHOT_ENTRIES=5

SNAPSHOT_FORMAT_CHOICE=5
`, cmdline)

	if err := p.fs.WriteFile(config.PathMntEtcDefaultLimine, []byte(limineDefaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write limine config: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Limine snapper config created")

	// Enable service
	if err := p.chrExec.ChrootSystemctl(p.logger.LogPath(), config.PathMnt, "enable", "limine-snapper-sync.service"); err != nil {
		return fmt.Errorf("failed to enable limine-snapper-sync: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Btrfs snapshots configured")
	return nil
}

// configurePacman configures pacman for better UX
func (p *PostInstallPhase) configurePacman(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring pacman UX...")

	content, err := p.fs.ReadFile(config.PathMntEtcPacmanConf)
	if err != nil {
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}

	contentStr := string(content)

	// Enable Color
	colorRegex := regexp.MustCompile(`(?m)^#Color$`)
	contentStr = colorRegex.ReplaceAllString(contentStr, "Color")

	// Enable ParallelDownloads
	parallelRegex := regexp.MustCompile(`(?m)^#ParallelDownloads = 5$`)
	contentStr = parallelRegex.ReplaceAllString(contentStr, "ParallelDownloads = 5")

	// Add ILoveCandy after Color
	switch {
	case !strings.Contains(contentStr, "ILoveCandy"):
		colorLineRegex := regexp.MustCompile(`(?m)^Color$`)
		contentStr = colorLineRegex.ReplaceAllString(contentStr, "Color\nILoveCandy")
	}

	if err := p.fs.WriteFile(config.PathMntEtcPacmanConf, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write pacman.conf: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Pacman configured (color, parallel downloads, candy)")
	return nil
}

// installPacmanHooks installs pacman hooks for maintenance
func (p *PostInstallPhase) installPacmanHooks(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Installing pacman hooks...")

	// Create hooks directory
	if err := p.fs.MkdirAll(config.PathMntEtcPacmanDHooks, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Limine update hook
	limineHook := `[Trigger]
Operation = Install
Operation = Upgrade
Type = Package
Target = limine

[Action]
Description = Deploying Limine after upgrade...
When = PostTransaction
Exec = /usr/bin/cp /usr/share/limine/BOOTX64.EFI /boot/EFI/limine/
`

	hookPath := filepath.Join(config.PathMntEtcPacmanDHooks, "99-limine.hook")
	if err := p.fs.WriteFile(hookPath, []byte(limineHook), 0644); err != nil {
		return fmt.Errorf("failed to write limine hook: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Limine update hook installed")
	return nil
}

// verifyInstallation performs comprehensive verification
func (p *PostInstallPhase) verifyInstallation(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Verifying installation components...")

	verificationFailed := false

	// Verify kernel
	kernelName := p.config.KernelChoice
	kernelPath := filepath.Join(config.PathMntBoot, fmt.Sprintf("vmlinuz-%s", kernelName))
	switch _, err := p.fs.Stat(kernelPath); {
	case p.fs.IsNotExist(err):
		p.SendOutput(progressChan, fmt.Sprintf("[ERROR] Kernel missing: %s", kernelPath))
		verificationFailed = true
	default:
		p.SendOutput(progressChan, "[OK] Kernel found")
	}

	// Verify initramfs
	initramfsPath := filepath.Join(config.PathMntBoot, fmt.Sprintf("initramfs-%s.img", kernelName))
	switch _, err := p.fs.Stat(initramfsPath); {
	case p.fs.IsNotExist(err):
		p.SendOutput(progressChan, fmt.Sprintf("[ERROR] Initramfs missing: %s", initramfsPath))
		verificationFailed = true
	default:
		p.SendOutput(progressChan, "[OK] Initramfs found")
	}

	// Verify bootloader
	bootloaderPath := filepath.Join(config.PathMntBootEFILimine, config.FileLimineBootloader)
	switch _, err := p.fs.Stat(bootloaderPath); {
	case p.fs.IsNotExist(err):
		p.SendOutput(progressChan, "[ERROR] Bootloader missing")
		verificationFailed = true
	default:
		p.SendOutput(progressChan, "[OK] Bootloader found")
	}

	// Verify Limine config
	limineConfPath := filepath.Join(config.PathMntBootEFILimine, config.FileLimineConfig)
	switch content, err := p.fs.ReadFile(limineConfPath); {
	case err != nil:
		p.SendOutput(progressChan, "[ERROR] Limine config missing")
		verificationFailed = true
	default:
		contentStr := string(content)
		switch {
		case strings.Contains(contentStr, "protocol:") && strings.Contains(contentStr, "path:"):
			p.SendOutput(progressChan, "[OK] Limine config validated")
		default:
			p.SendOutput(progressChan, "[WARN] Limine config may be incomplete")
		}
	}

	// Verify NetworkManager is enabled
	result := p.sysExec.RunSimple("arch-chroot", config.PathMnt, "systemctl", "is-enabled", "NetworkManager")
	switch {
	case result.ExitCode != 0:
		p.SendOutput(progressChan, "[ERROR] NetworkManager not enabled")
		verificationFailed = true
	default:
		p.SendOutput(progressChan, "[OK] NetworkManager enabled")
	}

	switch {
	case verificationFailed:
		p.SendOutput(progressChan, "[WARN] Verification completed with warnings")
		return fmt.Errorf("verification completed with warnings")
	default:
		p.SendOutput(progressChan, "[OK] All critical components verified")
	}

	return nil
}

// unmountAll unmounts all filesystems
func (p *PostInstallPhase) unmountAll(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Unmounting all partitions...")

	// Unmount /mnt recursively
	if err := system.Unmount(p.logger.LogPath(), config.PathMnt); err != nil {
		p.SendOutput(progressChan, "[WARN] Some partitions may still be mounted")
		return fmt.Errorf("failed to unmount cleanly: %w", err)
	}

	p.SendOutput(progressChan, "[OK] All partitions unmounted")

	// Close encrypted volume if encryption was used
	switch p.config.EncryptionType {
	case config.EncryptionLUKS, config.EncryptionLUKSLVM:
		p.SendOutput(progressChan, "Closing encrypted volume...")
		result := p.logger.ExecCommand("cryptsetup", "close", "cryptroot")
		switch {
		case result.ExitCode != 0:
			p.SendOutput(progressChan, "[WARN] Failed to close encrypted volume")
		default:
			p.SendOutput(progressChan, "[OK] Encrypted volume closed")
		}
	}

	return nil
}

// PostCheck validates post-install completion
func (p *PostInstallPhase) PostCheck() error {
	// Verification already done in Execute
	return p.config.Save()
}

// Rollback for post-install phase
func (p *PostInstallPhase) Rollback() error {
	// Attempt to unmount if something failed
	_ = system.Unmount(p.logger.LogPath(), config.PathMnt)
	return nil
}

// CanSkip returns false - post-install cannot be skipped
func (p *PostInstallPhase) CanSkip() bool {
	return false
}

