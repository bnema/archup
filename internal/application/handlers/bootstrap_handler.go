package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/domain/ports"
)

// BootstrapHandler handles downloading/cloning install files
type BootstrapHandler struct {
	fs         ports.FileSystem
	httpClient ports.HTTPClient
	logger     ports.Logger
	repoURL    string
	rawURL     string
	branch     string
}

// NewBootstrapHandler creates a new bootstrap handler
func NewBootstrapHandler(fs ports.FileSystem, httpClient ports.HTTPClient, logger ports.Logger, repoURL, rawURL, branch string) *BootstrapHandler {
	return &BootstrapHandler{
		fs:         fs,
		httpClient: httpClient,
		logger:     logger,
		repoURL:    repoURL,
		rawURL:     rawURL,
		branch:     branch,
	}
}

// Handle prepares install files by cloning repo or downloading
func (h *BootstrapHandler) Handle(ctx context.Context) (*dto.BootstrapResult, error) {
	h.logger.Info("Starting bootstrap")

	result := &dto.BootstrapResult{
		Success:     false,
		InstallDir:  config.DefaultInstallDir,
		ErrorDetail: "",
	}

	// Create install directory
	if err := h.fs.MkdirAll(config.DefaultInstallDir, 0755); err != nil {
		result.ErrorDetail = fmt.Sprintf("Failed to create install directory: %v", err)
		return result, err
	}

	// Try to clone repo first
	if err := h.cloneRepo(ctx); err == nil {
		if err := h.copyRepoFiles(); err != nil {
			h.logger.Warn("Failed to copy repo files", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to copy repo files: %v", err)
			return result, err
		}
		h.logger.Info("Bootstrap complete via repo clone")
		result.Success = true
		result.Method = "clone"
		return result, nil
	} else {
		h.logger.Warn("Repo clone failed, falling back to downloads", "error", err)
	}

	// Fallback: download files
	if err := h.downloadFiles(ctx); err != nil {
		result.ErrorDetail = fmt.Sprintf("Failed to download files: %v", err)
		return result, err
	}

	h.logger.Info("Bootstrap complete via downloads")
	result.Success = true
	result.Method = "download"
	return result, nil
}

// cloneRepo clones the archup repository
func (h *BootstrapHandler) cloneRepo(ctx context.Context) error {
	repoDir := config.DefaultInstallRepoDir

	// Check if already cloned
	if _, err := h.fs.Stat(filepath.Join(repoDir, ".git")); err == nil {
		h.logger.Info("Repository already cloned")
		return nil
	}

	h.logger.Info("Cloning repository", "url", h.repoURL)

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", h.branch, h.repoURL, repoDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// copyRepoFiles copies install files from cloned repo to install directory
func (h *BootstrapHandler) copyRepoFiles() error {
	repoDir := config.DefaultInstallRepoDir
	installDir := config.DefaultInstallDir

	// Files to copy from repo
	files := []struct {
		src  string
		dest string
	}{
		{"install/base.packages", "base.packages"},
		{"install/extra.packages", "extra.packages"},
		{"install/configs/limine.conf.template", "configs/limine.conf.template"},
		{"install/configs/chaotic-aur.conf", "configs/chaotic-aur.conf"},
		{"install/configs/shell/shell", "configs/shell/shell"},
		{"install/configs/shell/starship.toml", "configs/shell/starship.toml"},
		{"install/configs/shell/init", "configs/shell/init"},
		{"install/configs/shell/aliases", "configs/shell/aliases"},
		{"install/configs/shell/envs", "configs/shell/envs"},
		{"install/configs/shell/rc", "configs/shell/rc"},
		{"install/configs/shell/bashrc", "configs/shell/bashrc"},
	}

	for _, f := range files {
		srcPath := filepath.Join(repoDir, f.src)
		destPath := filepath.Join(installDir, f.dest)

		// Create destination directory
		if err := h.fs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", f.dest, err)
		}

		// Read and copy file
		content, err := h.fs.ReadFile(srcPath)
		if err != nil {
			h.logger.Warn("File not found in repo", "file", f.src)
			continue // Skip missing files
		}

		if err := h.fs.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.dest, err)
		}

		h.logger.Info("Copied file", "file", f.dest)
	}

	return nil
}

// downloadFiles downloads install files from GitHub
func (h *BootstrapHandler) downloadFiles(_ context.Context) error {
	files := []struct {
		urlPath  string
		destPath string
	}{
		{"install/base.packages", "base.packages"},
		{"install/extra.packages", "extra.packages"},
		{"install/configs/limine.conf.template", "configs/limine.conf.template"},
	}

	for _, f := range files {
		url := fmt.Sprintf("%s/%s", h.rawURL, f.urlPath)
		destPath := filepath.Join(config.DefaultInstallDir, f.destPath)

		// Create destination directory
		if err := h.fs.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", f.destPath, err)
		}

		h.logger.Info("Downloading file", "url", url)

		resp, err := h.httpClient.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", f.urlPath, err)
		}
		defer resp.Close()

		if err := h.fs.WriteFile(destPath, resp.Body(), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", f.destPath, err)
		}
	}

	return nil
}
