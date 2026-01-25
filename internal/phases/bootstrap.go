package phases

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

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
	if err != nil {
		return fmt.Errorf("no internet connectivity: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("failed to close connectivity response: %w", err)
	}

	return nil
}

// Execute runs the bootstrap phase
func (p *BootstrapPhase) Execute(progressChan chan<- ProgressUpdate) PhaseResult {
	p.SendOutput(progressChan, "Starting bootstrap phase...")

	// Create install directory
	if err := p.fs.MkdirAll(config.DefaultInstallDir, 0755); err != nil {
		p.SendError(progressChan, err)
		return PhaseResult{Success: false, Error: err}
	}

	if err := p.cloneRepo(); err == nil {
		if err := p.copyRepoFiles(progressChan); err != nil {
			p.SendError(progressChan, err)
			return PhaseResult{Success: false, Error: err}
		}
		p.SendComplete(progressChan, "Bootstrap complete")
		return PhaseResult{Success: true, Message: "Configuration files prepared from repo clone"}
	} else {
		p.SendOutput(progressChan, fmt.Sprintf("[WARN] Repo clone failed, falling back to downloads: %v", err))
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

		if err := p.downloadFile(file.url, file.destPath); err != nil {
			errMsg := fmt.Errorf("failed to download %s: %w", fileName, err)
			p.SendError(progressChan, errMsg)
			return PhaseResult{Success: false, Error: errMsg}
		}

		p.SendOutput(progressChan, fmt.Sprintf("[OK] %s", fileName))
	}

	p.SendComplete(progressChan, "Bootstrap complete")
	return PhaseResult{Success: true, Message: "All configuration files downloaded"}
}

func (p *BootstrapPhase) cloneRepo() error {
	if p.config.RepoURL == "" {
		return fmt.Errorf("repo URL not configured")
	}
	if err := p.fs.RemoveAll(config.DefaultInstallRepoDir); err != nil {
		return fmt.Errorf("failed to clear repo directory: %w", err)
	}

	branch := branchFromRawURL(p.config.RawURL)
	args := []string{"clone", "--depth", "1", "--branch", branch, "--single-branch", p.config.RepoURL, config.DefaultInstallRepoDir}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (p *BootstrapPhase) copyRepoFiles(progressChan chan<- ProgressUpdate) error {
	repoRoot := config.DefaultInstallRepoDir
	files := []struct {
		src  string
		dest string
	}{
		{filepath.Join(repoRoot, "install", "base.packages"), filepath.Join(config.DefaultInstallDir, "base.packages")},
		{filepath.Join(repoRoot, "install", "extra.packages"), filepath.Join(config.DefaultInstallDir, "extra.packages")},
		{filepath.Join(repoRoot, "install", "configs", "limine.conf.template"), filepath.Join(config.DefaultInstallDir, "configs", "limine.conf.template")},
		{filepath.Join(repoRoot, "install", "configs", "chaotic-aur.conf"), filepath.Join(config.DefaultInstallDir, "configs", "chaotic-aur.conf")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "shell"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "shell")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "starship.toml"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "starship.toml")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "init"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "init")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "aliases"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "aliases")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "envs"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "envs")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "rc"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "rc")},
		{filepath.Join(repoRoot, "install", "configs", "shell", "bashrc"), filepath.Join(config.DefaultInstallDir, "configs", "shell", "bashrc")},
		{filepath.Join(repoRoot, config.ArchLogoURL), filepath.Join(config.DefaultInstallDir, config.ArchLogoURL)},
	}

	for _, file := range config.PlymouthFiles {
		files = append(files, struct {
			src  string
			dest string
		}{
			src:  filepath.Join(repoRoot, "assets", "plymouth", file),
			dest: filepath.Join(config.DefaultInstallDir, "assets", "plymouth", file),
		})
	}

	for i, file := range files {
		fileName := filepath.Base(file.dest)
		p.SendProgress(progressChan, fmt.Sprintf("Preparing %s...", fileName), i+1, len(files))
		if err := p.copyFile(file.src, file.dest); err != nil {
			return fmt.Errorf("failed to prepare %s: %w", fileName, err)
		}
	}

	return nil
}

func (p *BootstrapPhase) copyFile(src, dest string) error {
	if err := p.fs.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	content, err := p.fs.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}
	if err := p.fs.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	return nil
}

func branchFromRawURL(rawURL string) string {
	const prefix = "https://raw.githubusercontent.com/"
	if !strings.HasPrefix(rawURL, prefix) {
		return "main"
	}
	trimmed := strings.TrimPrefix(rawURL, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 {
		return "main"
	}
	return parts[2]
}

// downloadFile downloads a file from URL to destPath
func (p *BootstrapPhase) downloadFile(url, destPath string) error {
	// Create destination directory if needed
	destDir := filepath.Dir(destPath)
	if err := p.fs.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download file
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return fmt.Errorf("failed to close response: %w", closeErr)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create destination file
	destFile, err := p.fs.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			p.logger.Warn("Failed to close file", "error", err)
		}
	}()

	// Copy content
	if _, err := io.Copy(destFile, resp.Body); err != nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return fmt.Errorf("failed to close response: %w", closeErr)
		}
		return fmt.Errorf("failed to write file: %w", err)
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		return fmt.Errorf("failed to close response: %w", closeErr)
	}

	return nil
}

// Rollback removes downloaded files
func (p *BootstrapPhase) Rollback() error {
	// Remove install directory
	if err := p.fs.RemoveAll(config.DefaultInstallDir); err != nil {
		return fmt.Errorf("failed to remove install directory: %w", err)
	}

	return nil
}
