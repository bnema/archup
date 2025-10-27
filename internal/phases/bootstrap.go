package phases

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// BootstrapPhase handles initial setup and file downloads
type BootstrapPhase struct {
	*BasePhase
	httpClient interfaces.HTTPClient
	fs         interfaces.FileSystem
}

// NewBootstrapPhase creates a new bootstrap phase
func NewBootstrapPhase(cfg *config.Config, log *logger.Logger, httpClient interfaces.HTTPClient, fs interfaces.FileSystem) *BootstrapPhase {
	return &BootstrapPhase{
		BasePhase:  NewBasePhase("bootstrap", "Bootstrap", cfg, log),
		httpClient: httpClient,
		fs:         fs,
	}
}

// PreCheck validates bootstrap prerequisites
func (p *BootstrapPhase) PreCheck() error {
	// Check internet connectivity
	resp, err := p.httpClient.Get("https://raw.githubusercontent.com")
	switch {
	case err != nil:
		return fmt.Errorf("no internet connectivity: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Execute runs the bootstrap phase
func (p *BootstrapPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendOutput(progressChan, "Starting bootstrap phase...")

	// Create install directory
	switch err := p.fs.MkdirAll(config.DefaultInstallDir, 0755); {
	case err != nil:
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	// Files to download
	files := []struct {
		url      string
		destPath string
	}{
		// Package lists
		{
			url:      fmt.Sprintf("%s/install/base.packages", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "base.packages"),
		},
		{
			url:      fmt.Sprintf("%s/install/extra.packages", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "extra.packages"),
		},
		// Config files
		{
			url:      fmt.Sprintf("%s/install/configs/limine.conf.template", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "limine.conf.template"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/chaotic-aur.conf", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "chaotic-aur.conf"),
		},
		// Shell config files
		{
			url:      fmt.Sprintf("%s/install/configs/shell/shell", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "shell"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/starship.toml", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "starship.toml"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/init", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "init"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/aliases", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "aliases"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/envs", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "envs"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/rc", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "rc"),
		},
		{
			url:      fmt.Sprintf("%s/install/configs/shell/bashrc", p.config.RawURL),
			destPath: filepath.Join(config.DefaultInstallDir, "configs", "shell", "bashrc"),
		},
	}

	totalFiles := len(files)
	p.SendOutput(progressChan, fmt.Sprintf("Downloading %d configuration files...", totalFiles))

	for i, file := range files {
		fileName := filepath.Base(file.destPath)
		p.SendProgress(progressChan, fmt.Sprintf("Downloading %s...", fileName), i+1, totalFiles)

		switch err := p.downloadFile(file.url, file.destPath); {
		case err != nil:
			errMsg := fmt.Errorf("failed to download %s: %w", fileName, err)
			p.SendError(progressChan, errMsg)
			return PhaseResult{Success: false, Error: errMsg}
		}

		p.SendOutput(progressChan, fmt.Sprintf("[OK] %s", fileName))
	}

	p.SendComplete(progressChan, "Bootstrap complete")
	return PhaseResult{Success: true, Message: "All configuration files downloaded"}
}

// downloadFile downloads a file from URL to destPath
func (p *BootstrapPhase) downloadFile(url, destPath string) error {
	// Create destination directory if needed
	destDir := filepath.Dir(destPath)
	switch err := p.fs.MkdirAll(destDir, 0755); {
	case err != nil:
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download file
	resp, err := p.httpClient.Get(url)
	switch {
	case err != nil:
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create destination file
	destFile, err := p.fs.Create(destPath)
	switch {
	case err != nil:
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	switch _, err := io.Copy(destFile, resp.Body); {
	case err != nil:
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Rollback removes downloaded files
func (p *BootstrapPhase) Rollback() error {
	// Remove install directory
	switch err := p.fs.RemoveAll(config.DefaultInstallDir); {
	case err != nil:
		return fmt.Errorf("failed to remove install directory: %w", err)
	}

	return nil
}
