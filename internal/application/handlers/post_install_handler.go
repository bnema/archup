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

		result.TasksRun = append(result.TasksRun, fmt.Sprintf("plymouth-theme-%s", cmd.PlymouthTheme))
		h.logger.Info("Plymouth theme installed")
	}

	if err := h.installLimineLogo(cmd.MountPoint); err != nil {
		h.logger.Warn("Failed to install Limine logo", "error", err)
	}

	// Final cleanup and verification
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
		if err := h.fs.Chmod(dst, 0755); err != nil {
			return fmt.Errorf("failed to set permissions for %s: %w", script, err)
		}
	}

	if err := h.chrExec.ChrootSystemctl(ctx, h.logger.LogPath(), mountPoint, "enable", config.PostBootServiceName); err != nil {
		return fmt.Errorf("failed to enable first-boot service: %w", err)
	}

	return nil
}

func (h *PostInstallHandler) installLimineLogo(mountPoint string) error {
	logoContent, err := h.downloadTemplate(config.ArchLogoURL)
	if err != nil {
		return err
	}

	logoPath := filepath.Join(mountPoint, "boot", "arch-logo.png")
	if err := h.fs.WriteFile(logoPath, logoContent, 0644); err != nil {
		return fmt.Errorf("failed to write Limine logo: %w", err)
	}

	limineConf := filepath.Join(mountPoint, "boot", "limine.conf")
	content, err := h.fs.ReadFile(limineConf)
	if err != nil {
		return fmt.Errorf("failed to read limine.conf: %w", err)
	}

	contentStr := string(content)
	graphicsRegex := regexp.MustCompile(`(?m)^graphics: yes$`)
	if graphicsRegex.MatchString(contentStr) {
		wallpaperSettings := "\nwallpaper: boot():/arch-logo.png\nwallpaper_style: centered\nbackdrop: 000000"
		contentStr = graphicsRegex.ReplaceAllString(contentStr, "graphics: yes"+wallpaperSettings)
		if err := h.fs.WriteFile(limineConf, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to update limine.conf: %w", err)
		}
		return nil
	}

	h.logger.Warn("graphics: yes not found in limine.conf")
	return nil
}

// copyShellConfigs is a no-op: shell configuration is handled by
// cli-tools.sh on first boot. No config files are copied at install time.
func (h *PostInstallHandler) copyShellConfigs(mountPoint, username string) error {
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
