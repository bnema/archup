package postboot

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/interfaces"
	"github.com/bnema/archup/internal/logger"
)

// fetchFile returns the content of a repo-relative path.
// It checks /tmp/archup-install/<repoPath> first (populated by push-binary-to-vm.sh --local),
// then falls back to downloading from GitHub.
func (s *Setup) fetchFile(repoPath string) ([]byte, error) {
	localPath := filepath.Join(config.DefaultInstallDir, repoPath)
	if content, err := s.fs.ReadFile(localPath); err == nil {
		return content, nil
	}
	url := fmt.Sprintf("%s/%s", s.config.RawURL, repoPath)
	resp, err := s.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", repoPath, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch %s: HTTP %d", repoPath, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", repoPath, err)
	}
	return data, nil
}

// SetupResult holds the result of post-boot setup
type SetupResult struct {
	ScriptsDownloaded int
	ServiceEnabled    bool
}

// Setup handles first-boot service configuration
type Setup struct {
	fs      interfaces.FileSystem
	http    interfaces.HTTPClient
	chrExec interfaces.ChrootExecutor
	config  *config.Config
	logger  *logger.Logger
}

// NewSetup creates a new post-boot setup handler
func NewSetup(
	fs interfaces.FileSystem,
	http interfaces.HTTPClient,
	chrExec interfaces.ChrootExecutor,
	cfg *config.Config,
	log *logger.Logger,
) *Setup {
	return &Setup{
		fs:      fs,
		http:    http,
		chrExec: chrExec,
		config:  cfg,
		logger:  log,
	}
}

// Configure downloads and configures first-boot service
func (s *Setup) Configure() (SetupResult, error) {
	result := SetupResult{
		ScriptsDownloaded: 0,
		ServiceEnabled:    false,
	}

	if err := s.fs.MkdirAll(config.PathMntPostBoot, 0755); err != nil {
		return result, fmt.Errorf("failed to create post-boot directory: %w", err)
	}

	// Fetch logo.txt (local-first, then remote)
	logoData, err := s.fetchFile("logo.txt")
	if err != nil {
		return result, fmt.Errorf("failed to fetch logo.txt: %w", err)
	}
	logoPath := filepath.Join(config.PathMntPostBoot, "logo.txt")
	if err := s.fs.WriteFile(logoPath, logoData, 0644); err != nil {
		return result, fmt.Errorf("failed to write logo.txt: %w", err)
	}

	// Fetch all post-boot scripts (local-first, then remote)
	for _, script := range config.PostBootScripts {
		data, err := s.fetchFile(filepath.Join("install", "mandatory", "post-boot", script))
		if err != nil {
			return result, fmt.Errorf("failed to fetch %s: %w", script, err)
		}

		scriptPath := filepath.Join(config.PathMntPostBoot, script)
		if err := s.fs.WriteFile(scriptPath, data, 0755); err != nil {
			return result, fmt.Errorf("failed to write %s: %w", script, err)
		}

		if err := s.fs.Chmod(scriptPath, 0755); err != nil {
			return result, fmt.Errorf("failed to set permissions on %s: %w", script, err)
		}
	}

	// Fetch service template (local-first, then remote)
	templateBytes, err := s.fetchFile(config.PostBootServiceTemplate)
	if err != nil {
		return result, fmt.Errorf("failed to fetch service template: %w", err)
	}

	serviceContent := string(templateBytes)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_EMAIL__", s.config.Email)
	serviceContent = strings.ReplaceAll(serviceContent, "__ARCHUP_USERNAME__", s.config.Username)

	servicePath := filepath.Join(config.PathMntSystemdSystem, config.PostBootServiceName)
	if err := s.fs.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return result, fmt.Errorf("failed to write service file: %w", err)
	}

	if err := s.chrExec.ChrootSystemctl(s.logger.LogPath(), config.PathMnt, "enable", config.PostBootServiceName); err != nil {
		return result, fmt.Errorf("failed to enable service: %w", err)
	}

	result.ServiceEnabled = true
	result.ScriptsDownloaded = len(config.PostBootScripts)
	return result, nil
}
