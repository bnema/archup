package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/domain/ports"
)

// PostInstallHandler handles post-installation tasks
type PostInstallHandler struct {
	fs         ports.FileSystem
	httpClient ports.HTTPClient
	chrExec    ports.ChrootExecutor
	scriptExec ports.ScriptExecutor
	logger     ports.Logger
	rawURL     string
}

// NewPostInstallHandler creates a new post-installation handler
func NewPostInstallHandler(fs ports.FileSystem, httpClient ports.HTTPClient, chrExec ports.ChrootExecutor, scriptExec ports.ScriptExecutor, logger ports.Logger, rawURL string) *PostInstallHandler {
	return &PostInstallHandler{
		fs:         fs,
		httpClient: httpClient,
		chrExec:    chrExec,
		scriptExec: scriptExec,
		logger:     logger,
		rawURL:     rawURL,
	}
}

// Handle executes post-installation tasks
func (h *PostInstallHandler) Handle(ctx context.Context, cmd commands.PostInstallCommand) (*dto.PostInstallResult, error) {
	h.logger.Info("Starting post-installation tasks", "username", cmd.Username)

	result := &dto.PostInstallResult{
		Success:     false,
		TasksRun:    []string{},
		ErrorDetail: "",
	}

	// Run post-boot scripts if requested
	if cmd.RunPostBootScripts {
		h.logger.Info("Running post-boot scripts")
		result.TasksRun = append(result.TasksRun, "post-boot-scripts")

		if err := h.setupPostBoot(ctx, cmd.MountPoint, cmd.Username, cmd.UserEmail); err != nil {
			h.logger.Error("Failed to setup post-boot scripts", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to setup post-boot scripts: %v", err)
			return result, err
		}

		h.logger.Info("Post-boot scripts prepared")
	}

	// Write Dank Linux flag file if user opted in
	if cmd.InstallDankLinux {
		if err := h.writeDankLinuxFlag(cmd.MountPoint); err != nil {
			result.ErrorDetail = fmt.Sprintf("Failed to write Dank Linux flag: %v", err)
			return result, fmt.Errorf("failed to write dank linux flag: %w", err)
		}
		result.TasksRun = append(result.TasksRun, "dank-linux-flag")
		h.logger.Info("Dank Linux flag file written")
	}

	// Setup limine-snapper-sync for btrfs snapshot bootability
	if err := h.setupSnapperSync(ctx, cmd.MountPoint); err != nil {
		h.logger.Warn("Failed to setup limine-snapper-sync", "error", err)
	}

	// Install Plymouth theme if specified
	if cmd.PlymouthTheme != "" {
		h.logger.Info("Installing Plymouth theme", "theme", cmd.PlymouthTheme)

		themeDir := filepath.Join(cmd.MountPoint, "usr", "share", "plymouth", "themes", cmd.PlymouthTheme)
		if err := h.fs.MkdirAll(themeDir, 0755); err != nil {
			return result, fmt.Errorf("failed to create Plymouth theme directory: %w", err)
		}

		for _, file := range config.PlymouthFiles {
			content, err := h.downloadTemplate(filepath.Join("assets", "plymouth", file))
			if err != nil {
				return result, fmt.Errorf("failed to download Plymouth file %s: %w", file, err)
			}
			if err := h.fs.WriteFile(filepath.Join(themeDir, file), content, 0644); err != nil {
				return result, fmt.Errorf("failed to write Plymouth file %s: %w", file, err)
			}
		}

		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "plymouth-set-default-theme", cmd.PlymouthTheme); err != nil {
			return result, fmt.Errorf("failed to set Plymouth theme: %w", err)
		}

		// Rebuild initramfs so the Plymouth theme takes effect at boot
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "mkinitcpio", "-P"); err != nil {
			return result, fmt.Errorf("failed to rebuild initramfs after Plymouth theme: %w", err)
		}

		result.TasksRun = append(result.TasksRun, fmt.Sprintf("plymouth-theme-%s", cmd.PlymouthTheme))
		h.logger.Info("Plymouth theme installed and initramfs rebuilt")
	}

	if err := h.tunePacmanConfig(cmd.MountPoint); err != nil {
		h.logger.Warn("Failed to tune pacman.conf", "error", err)
	}

	if err := h.installLimineHook(cmd.MountPoint, cmd.TargetDisk); err != nil {
		h.logger.Warn("Failed to install limine hook", "error", err)
	}

	// Final cleanup and verification
	result.VerificationWarnings = h.verifyInstallation(cmd.MountPoint, cmd.Encrypted)
	if len(result.VerificationWarnings) > 0 {
		h.logger.Warn("Post-install verification warnings", "warnings", result.VerificationWarnings)
	}

	h.logger.Info("Post-installation tasks completed")
	result.Success = true

	return result, nil
}

func (h *PostInstallHandler) setupPostBoot(ctx context.Context, mountPoint, username, email string) error {
	postBootPath := filepath.Join(mountPoint, "usr", "local", "share", "archup", "post-boot")
	if err := h.fs.MkdirAll(postBootPath, 0755); err != nil {
		return fmt.Errorf("failed to create post-boot directory: %w", err)
	}

	// Write service file with placeholder replacement for email and username
	serviceDest := filepath.Join(mountPoint, "etc", "systemd", "system", "archup-first-boot.service")
	if err := h.writeServiceFile("install/mandatory/post-boot/archup-first-boot.service", serviceDest, username, email); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	for _, script := range config.PostBootScripts {
		src := filepath.Join("install", "mandatory", "post-boot", script)
		dst := filepath.Join(postBootPath, script)
		if err := h.writeFromTemplate(src, dst); err != nil {
			return fmt.Errorf("failed to write %s: %w", script, err)
		}
		// Only set executable bit on shell scripts
		if strings.HasSuffix(script, ".sh") {
			if err := h.fs.Chmod(dst, 0755); err != nil {
				return fmt.Errorf("failed to set permissions for %s: %w", script, err)
			}
		}
	}

	if err := h.chrExec.ChrootSystemctl(ctx, h.logger.LogPath(), mountPoint, "enable", config.PostBootServiceName); err != nil {
		return fmt.Errorf("failed to enable first-boot service: %w", err)
	}

	return nil
}

func (h *PostInstallHandler) writeFromTemplate(src string, dst string) error {
	content, err := h.tryReadLocal(src)
	if err != nil {
		content, err = h.downloadTemplate(src)
		if err != nil {
			return err
		}
	}

	return h.fs.WriteFile(dst, content, 0644)
}

// writeServiceFile writes the systemd service file with placeholder replacement
func (h *PostInstallHandler) writeServiceFile(src, dst, username, email string) error {
	content, err := h.tryReadLocal(src)
	if err != nil {
		content, err = h.downloadTemplate(src)
		if err != nil {
			return err
		}
	}

	// Replace placeholders with actual values
	serviceContent := string(content)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_USERNAME__", username)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_EMAIL__", email)

	return h.fs.WriteFile(dst, []byte(serviceContent), 0644)
}

func (h *PostInstallHandler) tryReadLocal(path string) ([]byte, error) {
	localPath := filepath.Join(config.DefaultInstallDir, path)
	if exists, err := h.fs.Exists(localPath); err == nil && exists {
		return h.fs.ReadFile(localPath)
	}
	if exists, err := h.fs.Exists(path); err == nil && exists {
		return h.fs.ReadFile(path)
	}
	return nil, fmt.Errorf("local template not found")
}

// writeDankLinuxFlag writes a flag file to the installed system so that
// dms-opt-in.sh auto-runs without prompting on first boot.
func (h *PostInstallHandler) writeDankLinuxFlag(mountPoint string) error {
	if err := h.fs.MkdirAll(filepath.Join(mountPoint, "var", "lib"), 0755); err != nil {
		return fmt.Errorf("failed to create /var/lib: %w", err)
	}
	flagPath := filepath.Join(mountPoint, "var", "lib", "archup-install-danklinux")
	return h.fs.WriteFile(flagPath, []byte(""), 0644)
}

func (h *PostInstallHandler) setupSnapperSync(ctx context.Context, mountPoint string) error {
	if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "pacman", "-S", "--noconfirm", "--needed", "limine-snapper-sync"); err != nil {
		return fmt.Errorf("failed to install limine-snapper-sync: %w", err)
	}
	if err := h.chrExec.ChrootSystemctl(ctx, h.logger.LogPath(), mountPoint, "enable", "limine-snapper-sync.service"); err != nil {
		return fmt.Errorf("failed to enable limine-snapper-sync.service: %w", err)
	}
	// Sanitize config: comment out enrollment commands when limine-entry-tool is absent
	if err := h.sanitizeLimineSnapperSyncConfig(mountPoint); err != nil {
		h.logger.Warn("Failed to sanitize limine-snapper-sync config", "error", err)
	}
	return nil
}

func (h *PostInstallHandler) sanitizeLimineSnapperSyncConfig(mountPoint string) error {
	confPath := filepath.Join(mountPoint, "etc", "limine-snapper-sync.conf")
	exists, err := h.fs.Exists(confPath)
	if err != nil || !exists {
		return err // config missing, nothing to sanitize
	}

	// If the binary is present, enrollment workflow is intentionally configured — leave config alone
	enrollBin := filepath.Join(mountPoint, "usr", "bin", "limine-reset-enroll")
	if _, err := h.fs.Stat(enrollBin); err == nil {
		return nil
	}

	b, err := h.fs.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to read limine-snapper-sync.conf: %w", err)
	}

	s := string(b)
	reBefore := regexp.MustCompile(`(?m)^COMMANDS_BEFORE_SAVE=.*$`)
	reAfter := regexp.MustCompile(`(?m)^COMMANDS_AFTER_SAVE=.*$`)
	s = reBefore.ReplaceAllString(s, `# COMMANDS_BEFORE_SAVE="" # disabled: limine-entry-tool not installed`)
	s = reAfter.ReplaceAllString(s, `# COMMANDS_AFTER_SAVE="" # disabled: limine-entry-tool not installed`)

	return h.fs.WriteFile(confPath, []byte(s), 0644)
}

// limineDiskPlaceholder is the placeholder in limine-update.hook replaced at install time.
const limineDiskPlaceholder = "DISK_PLACEHOLDER"

func (h *PostInstallHandler) installLimineHook(mountPoint, targetDisk string) error {
	hooksDir := filepath.Join(mountPoint, "etc", "pacman.d", "hooks")
	if err := h.fs.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks dir: %w", err)
	}
	content, err := h.tryReadLocal("install/configs/limine-update.hook")
	if err != nil {
		content, err = h.downloadTemplate("install/configs/limine-update.hook")
		if err != nil {
			return fmt.Errorf("failed to get limine hook template: %w", err)
		}
	}
	hookContent := strings.ReplaceAll(string(content), limineDiskPlaceholder, targetDisk)
	return h.fs.WriteFile(filepath.Join(hooksDir, "limine-update.hook"), []byte(hookContent), 0644)
}

func (h *PostInstallHandler) tunePacmanConfig(mountPoint string) error {
	confPath := filepath.Join(mountPoint, "etc", "pacman.conf")
	content, err := h.fs.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}
	conf := string(content)
	conf = uncommentPacmanOption(conf, "Color")
	conf = uncommentPacmanOption(conf, "ParallelDownloads")
	conf = uncommentPacmanOption(conf, "ILoveCandy")
	return h.fs.WriteFile(confPath, []byte(conf), 0644)
}

// uncommentPacmanOption uncomments a pacman.conf option, handling optional spaces after #.
func uncommentPacmanOption(conf, option string) string {
	re := regexp.MustCompile(`(?m)^#\s*(` + regexp.QuoteMeta(option) + `.*)$`)
	return re.ReplaceAllString(conf, "$1")
}

func (h *PostInstallHandler) verifyInstallation(mountPoint string, encrypted bool) []string {
	warnings := []string{}
	checks := []struct{ path, name string }{
		{filepath.Join(mountPoint, "etc", "fstab"), "fstab"},
		{filepath.Join(mountPoint, "boot", "limine.conf"), "limine.conf"},
		{filepath.Join(mountPoint, "boot", "EFI", "BOOT", "BOOTX64.EFI"), "EFI boot file"},
	}
	if encrypted {
		checks = append(checks, struct{ path, name string }{filepath.Join(mountPoint, "etc", "crypttab"), "crypttab"})
	}
	for _, c := range checks {
		if _, err := h.fs.Stat(c.path); err != nil {
			warnings = append(warnings, fmt.Sprintf("missing %s: %s", c.name, c.path))
		}
	}
	return warnings
}

func (h *PostInstallHandler) downloadTemplate(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", h.rawURL, path)
	h.logger.Info("Downloading template", "url", url)
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Close(); closeErr != nil {
			h.logger.Warn("Failed to close HTTP response", "error", closeErr)
		}
	}()
	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode(), url)
	}
	return resp.Body(), nil
}
