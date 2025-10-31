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

	// Download logo.txt
	logoURL := fmt.Sprintf("%s/logo.txt", s.config.RawURL)
	resp, err := s.http.Get(logoURL)
	switch {
	case err != nil:
		return result, fmt.Errorf("failed to download logo.txt: %w", err)
	case resp.StatusCode != http.StatusOK:
		resp.Body.Close()
		return result, fmt.Errorf("failed to download logo.txt: HTTP %d", resp.StatusCode)
	}

	logoPath := filepath.Join(config.PathMntPostBoot, "logo.txt")
	logoFile, err := s.fs.Create(logoPath)
	if err != nil {
		resp.Body.Close()
		return result, fmt.Errorf("failed to create logo.txt: %w", err)
	}

	_, err = io.Copy(logoFile, resp.Body)
	logoFile.Close()
	resp.Body.Close()
	if err != nil {
		return result, fmt.Errorf("failed to save logo.txt: %w", err)
	}

	// Download all post-boot scripts
	for _, script := range config.PostBootScripts {

		scriptURL := fmt.Sprintf("%s/install/post-boot/%s", s.config.RawURL, script)
		resp, err := s.http.Get(scriptURL)
		switch {
		case err != nil:
			return result, fmt.Errorf("failed to download %s: %w", script, err)
		case resp.StatusCode != http.StatusOK:
			resp.Body.Close()
			return result, fmt.Errorf("failed to download %s: HTTP %d", script, resp.StatusCode)
		}

		scriptPath := filepath.Join(config.PathMntPostBoot, script)
		scriptFile, err := s.fs.Create(scriptPath)
		if err != nil {
			resp.Body.Close()
			return result, fmt.Errorf("failed to create %s: %w", script, err)
		}

		_, err = io.Copy(scriptFile, resp.Body)
		scriptFile.Close()
		resp.Body.Close()

		if err != nil {
			return result, fmt.Errorf("failed to save %s: %w", script, err)
		}

		if err := s.fs.Chmod(scriptPath, 0755); err != nil {
			return result, fmt.Errorf("failed to set permissions on %s: %w", script, err)
		}
	}


	// Download service template
	serviceURL := fmt.Sprintf("%s/%s", s.config.RawURL, config.PostBootServiceTemplate)
	resp, err = s.http.Get(serviceURL)
	switch {
	case err != nil:
		return result, fmt.Errorf("failed to download service template: %w", err)
	case resp.StatusCode != http.StatusOK:
		resp.Body.Close()
		return result, fmt.Errorf("failed to download service template: HTTP %d", resp.StatusCode)
	}

	templateBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return result, fmt.Errorf("failed to read service template: %w", err)
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
