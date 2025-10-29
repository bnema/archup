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
	"github.com/bnema/archup/internal/system"
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
	totalSteps := 9
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
	if err := p.setupPostBoot(progressChan); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
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
	if err := p.configureShell(progressChan); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	// Step 8: Verification
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

// configureShell sets up bash shell with ArchUp defaults
func (p *PostInstallPhase) configureShell(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring shell environment...")

	// Get username from /mnt/home
	homeDir := filepath.Join(config.PathMnt, "home")
	entries, err := os.ReadDir(homeDir)
	switch {
	case err != nil:
		return fmt.Errorf("failed to read home directory: %w", err)
	case len(entries) == 0:
		return fmt.Errorf("no user found in /home")
	}

	username := entries[0].Name()
	userHome := filepath.Join(config.PathMnt, "home", username)
	archupDefault := filepath.Join(userHome, ".local", "share", "archup", "default")
	archupDefaultBash := filepath.Join(archupDefault, "bash")

	p.SendOutput(progressChan, fmt.Sprintf("Configuring shell for user: %s", username))

	// Create directory structure
	if err := p.fs.MkdirAll(archupDefaultBash, 0755); err != nil {
		return fmt.Errorf("failed to create shell config directory: %w", err)
	}

	// Copy shell configuration files from templates
	if err := p.copyShellTemplates(archupDefault, archupDefaultBash); err != nil {
		return fmt.Errorf("failed to copy shell templates: %w", err)
	}

	// Create .bashrc
	bashrcContent, err := p.readTemplate(config.BashrcTemplate)
	if err != nil {
		return fmt.Errorf("failed to read bashrc template: %w", err)
	}

	bashrcPath := filepath.Join(userHome, ".bashrc")
	if err := p.fs.WriteFile(bashrcPath, bashrcContent, 0644); err != nil {
		return fmt.Errorf("failed to write .bashrc: %w", err)
	}

	// Set ownership to user
	if err := p.setShellOwnership(userHome, username); err != nil {
		return fmt.Errorf("failed to set ownership: %w", err)
	}

	// Clone bleu-theme repository
	p.SendOutput(progressChan, "Cloning bleu-theme repository...")
	themesDir := filepath.Join(userHome, ".local", "share", "archup", "themes")
	bleuThemeDir := filepath.Join(themesDir, "bleu")
	currentDir := filepath.Join(userHome, ".local", "share", "archup", "current")

	cloneCmd := fmt.Sprintf("su - %s -c 'mkdir -p %s && git clone https://github.com/bnema/bleu-theme.git %s'",
		username, themesDir, bleuThemeDir)
	switch err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, cloneCmd); {
	case err != nil:
		p.SendOutput(progressChan, "[WARN] Failed to clone bleu-theme repository")
	default:
		p.SendOutput(progressChan, "[OK] Bleu-theme repository cloned")
	}

	// Create current theme symlink
	symlinkCmd := fmt.Sprintf("su - %s -c 'mkdir -p %s && ln -snf %s %s/theme'",
		username, currentDir, bleuThemeDir, currentDir)
	if err := p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, symlinkCmd); err != nil {
		p.SendOutput(progressChan, "[WARN] Failed to create theme symlink")
	}

	// Apply themes to CLI tools
	p.SendOutput(progressChan, "Applying themes to CLI tools...")
	if err := p.applyThemes(progressChan, username, userHome); err != nil {
		// Non-fatal, log and continue
		p.SendOutput(progressChan, "[WARN] Some themes failed to apply")
	}

	// Configure git delta
	gitCommands := []string{
		fmt.Sprintf("su - %s -c 'git config --global core.pager delta'", username),
		fmt.Sprintf("su - %s -c 'git config --global interactive.diffFilter \"delta --color-only\"'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.navigate true'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.light false'", username),
		fmt.Sprintf("su - %s -c 'git config --global delta.side-by-side true'", username),
		fmt.Sprintf("su - %s -c 'git config --global merge.conflictstyle diff3'", username),
		fmt.Sprintf("su - %s -c 'git config --global diff.colorMoved default'", username),
	}

	for _, gitCmd := range gitCommands {
		if err := p.chrExec.ChrootExec(p.logger.LogPath(),config.PathMnt, gitCmd); err != nil {
			// Non-fatal
			p.SendOutput(progressChan, "[WARN] Failed to configure git delta")
			break
		}
	}

	p.SendOutput(progressChan, "[OK] Shell environment configured")
	return nil
}

// copyShellTemplates copies shell config templates to user directory
func (p *PostInstallPhase) copyShellTemplates(archupDefault, archupDefaultBash string) error {
	templates := map[string]string{
		config.ShellConfigTemplate:  filepath.Join(archupDefaultBash, "shell"),
		config.ShellInitTemplate:    filepath.Join(archupDefaultBash, "init"),
		config.ShellAliasesTemplate: filepath.Join(archupDefaultBash, "aliases"),
		config.ShellEnvsTemplate:    filepath.Join(archupDefaultBash, "envs"),
		config.ShellRcTemplate:      filepath.Join(archupDefaultBash, "rc"),
		// Note: starship.toml is now copied from bleu-theme in applyThemes()
	}

	for template, dest := range templates {
		content, err := p.readTemplate(template)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", template, err)
		}

		if err := p.fs.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dest, err)
		}
	}

	return nil
}

// readTemplate reads a template file from install path
func (p *PostInstallPhase) readTemplate(filename string) ([]byte, error) {
	// Use DefaultInstallDir directly to match where bootstrap downloads files
	templatePath := filepath.Join(config.DefaultInstallDir, filename)
	return p.fs.ReadFile(templatePath)
}

// setShellOwnership sets ownership of shell config files to user
func (p *PostInstallPhase) setShellOwnership(userHome, username string) error {
	// Use chroot to run chown
	relativeHome := strings.TrimPrefix(userHome, config.PathMnt)
	chownCmd := fmt.Sprintf("chown -R %s:%s %s/.local %s/.bashrc",
		username, username, relativeHome, relativeHome)
	return p.chrExec.ChrootExec(p.logger.LogPath(),config.PathMnt, chownCmd)
}

// applyThemes applies bleu-theme to CLI tools
func (p *PostInstallPhase) applyThemes(progressChan chan<- ProgressUpdate, username, userHome string) error {
	themeDir := filepath.Join(userHome, ".local", "share", "archup", "current", "theme")
	archupDefault := filepath.Join(userHome, ".local", "share", "archup", "default")

	// Starship theme - copy to archup default location
	p.SendOutput(progressChan, "Configuring starship theme...")
	starshipCmd := fmt.Sprintf("su - %s -c 'cp %s/starship/bleu.toml %s/starship.toml'",
		username, themeDir, archupDefault)
	p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, starshipCmd)

	// Eza theme
	p.SendOutput(progressChan, "Configuring eza theme...")
	ezaConfigDir := fmt.Sprintf("%s/.config/eza", userHome)
	ezaCommands := []string{
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, ezaConfigDir),
		fmt.Sprintf("su - %s -c 'cp %s/eza/theme.yml %s/'", username, themeDir, ezaConfigDir),
	}
	for _, cmd := range ezaCommands {
		p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, cmd)
	}

	// Bat theme
	p.SendOutput(progressChan, "Configuring bat theme...")
	batThemeDir := fmt.Sprintf("%s/.config/bat/themes", userHome)
	batCommands := []string{
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, batThemeDir),
		fmt.Sprintf("su - %s -c 'cp %s/bat/bleu.tmTheme %s/'", username, themeDir, batThemeDir),
		fmt.Sprintf("su - %s -c 'bat cache --build'", username),
	}
	for _, cmd := range batCommands {
		p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, cmd)
	}

	// Btop theme
	p.SendOutput(progressChan, "Configuring btop theme...")
	btopThemeDir := fmt.Sprintf("%s/.config/btop/themes", userHome)
	btopCommands := []string{
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, btopThemeDir),
		fmt.Sprintf("su - %s -c 'cp %s/btop/bleu.theme %s/'", username, themeDir, btopThemeDir),
	}
	for _, cmd := range btopCommands {
		p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, cmd)
	}

	// Yazi theme
	p.SendOutput(progressChan, "Configuring yazi theme...")
	yaziFlavorDir := fmt.Sprintf("%s/.config/yazi/flavors/bleu", userHome)
	yaziCommands := []string{
		fmt.Sprintf("su - %s -c 'mkdir -p %s'", username, yaziFlavorDir),
		fmt.Sprintf("su - %s -c 'cp %s/yazi/bleu.toml %s/flavor.toml'", username, themeDir, yaziFlavorDir),
	}
	for _, cmd := range yaziCommands {
		p.chrExec.ChrootExec(p.logger.LogPath(), config.PathMnt, cmd)
	}

	p.SendOutput(progressChan, "[OK] CLI themes configured")
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

// setupPostBoot downloads and configures first-boot service
func (p *PostInstallPhase) setupPostBoot(progressChan chan<- ProgressUpdate) error {
	p.SendOutput(progressChan, "Configuring first-boot service...")

	if err := p.fs.MkdirAll(config.PathMntPostBoot, 0755); err != nil {
		return fmt.Errorf("failed to create post-boot directory: %w", err)
	}

	// Download logo.txt
	p.SendOutput(progressChan, "Downloading logo.txt...")
	logoURL := fmt.Sprintf("%s/logo.txt", p.config.RawURL)
	resp, err := p.http.Get(logoURL)
	switch {
	case err != nil:
		return fmt.Errorf("failed to download logo.txt: %w", err)
	case resp.StatusCode != http.StatusOK:
		resp.Body.Close()
		return fmt.Errorf("failed to download logo.txt: HTTP %d", resp.StatusCode)
	}

	logoPath := filepath.Join(config.PathMntPostBoot, "logo.txt")
	logoFile, err := p.fs.Create(logoPath)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to create logo.txt: %w", err)
	}

	_, err = io.Copy(logoFile, resp.Body)
	logoFile.Close()
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to save logo.txt: %w", err)
	}

	// Download all post-boot scripts
	for _, script := range config.PostBootScripts {
		p.SendOutput(progressChan, fmt.Sprintf("Downloading %s...", script))

		scriptURL := fmt.Sprintf("%s/install/post-boot/%s", p.config.RawURL, script)
		resp, err := p.http.Get(scriptURL)
		switch {
		case err != nil:
			return fmt.Errorf("failed to download %s: %w", script, err)
		case resp.StatusCode != http.StatusOK:
			resp.Body.Close()
			return fmt.Errorf("failed to download %s: HTTP %d", script, resp.StatusCode)
		}

		scriptPath := filepath.Join(config.PathMntPostBoot, script)
		scriptFile, err := p.fs.Create(scriptPath)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to create %s: %w", script, err)
		}

		_, err = io.Copy(scriptFile, resp.Body)
		scriptFile.Close()
		resp.Body.Close()

		if err != nil {
			return fmt.Errorf("failed to save %s: %w", script, err)
		}

		if err := p.fs.Chmod(scriptPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions on %s: %w", script, err)
		}
	}

	p.SendOutput(progressChan, "[OK] Post-boot scripts downloaded")

	// Download service template
	p.SendOutput(progressChan, "Creating systemd service...")
	serviceURL := fmt.Sprintf("%s/%s", p.config.RawURL, config.PostBootServiceTemplate)
	resp, err = p.http.Get(serviceURL)
	switch {
	case err != nil:
		return fmt.Errorf("failed to download service template: %w", err)
	case resp.StatusCode != http.StatusOK:
		resp.Body.Close()
		return fmt.Errorf("failed to download service template: HTTP %d", resp.StatusCode)
	}

	templateBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read service template: %w", err)
	}

	serviceContent := string(templateBytes)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_EMAIL__", p.config.Email)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_USERNAME__", p.config.Username)

	servicePath := filepath.Join(config.PathMntSystemdSystem, config.PostBootServiceName)
	if err := p.fs.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	p.SendOutput(progressChan, "[OK] Service file created")

	p.SendOutput(progressChan, "Enabling first-boot service...")
	if err := p.chrExec.ChrootSystemctl(p.logger.LogPath(), config.PathMnt, "enable", config.PostBootServiceName); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	p.SendOutput(progressChan, "[OK] First-boot service enabled")
	return nil
}
