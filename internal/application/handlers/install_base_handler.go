package handlers

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports"
)

// InstallBaseHandler handles base system installation
type InstallBaseHandler struct {
	fs      ports.FileSystem
	cmdExec ports.CommandExecutor
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewInstallBaseHandler creates a new base installation handler
func NewInstallBaseHandler(fs ports.FileSystem, cmdExec ports.CommandExecutor, chrExec ports.ChrootExecutor, logger ports.Logger) *InstallBaseHandler {
	return &InstallBaseHandler{
		fs:      fs,
		cmdExec: cmdExec,
		chrExec: chrExec,
		logger:  logger,
	}
}

// Handle installs the base system
func (h *InstallBaseHandler) Handle(ctx context.Context, cmd commands.InstallBaseCommand) (*dto.InstallBaseResult, error) {
	h.logger.Info("Starting base system installation", "kernel", cmd.KernelVariant)

	result := &dto.InstallBaseResult{
		Success:           false,
		PackagesInstalled: []string{},
		ErrorDetail:       "",
	}

	// Validate kernel variant
	kernel, err := packages.NewKernel(cmd.KernelVariant)
	if err != nil {
		h.logger.Error("Invalid kernel variant", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid kernel variant: %v", err)
		return result, err
	}

	// Load base packages from file (downloaded during bootstrap)
	basePackages, err := h.loadBasePackages()
	if err != nil {
		h.logger.Error("Failed to load base.packages", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to load base.packages: %v - run bootstrap first", err)
		return result, err
	}

	// Add kernel (not in base.packages, selected dynamically)
	basePackages = append(basePackages, kernel.PackageName())
	h.logger.Info("Adding kernel", "kernel", kernel.PackageName())

	// Add microcode if requested
	if cmd.IncludeMicrocode {
		basePackages = append(basePackages, "intel-ucode", "amd-ucode")
		h.logger.Info("Adding CPU microcode packages")
	}

	// Add cryptsetup for encrypted installs
	if cmd.Encrypted {
		basePackages = append(basePackages, "cryptsetup")
		h.logger.Info("Adding cryptsetup for encrypted install")
	}

	// Add any additional packages from command
	if len(cmd.Packages) > 0 {
		basePackages = append(basePackages, cmd.Packages...)
		h.logger.Info("Adding additional packages", "count", len(cmd.Packages))
	}

	// For CachyOS kernel: add the repo to the live host before pacstrap
	if cmd.KernelVariant == packages.KernelCachyOS {
		if err := h.setupCachyOSOnHost(ctx); err != nil {
			h.logger.Error("Failed to setup CachyOS repo on host", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to setup CachyOS repo on host: %v", err)
			return result, err
		}
	}

	h.logger.Info("Installing base packages", "count", len(basePackages))
	args := append([]string{cmd.MountPoint}, basePackages...)
	if _, err := h.cmdExec.Execute(ctx, "pacstrap", args...); err != nil {
		h.logger.Error("Pacstrap failed", "error", err)
		result.ErrorDetail = fmt.Sprintf("Pacstrap failed: %v", err)
		return result, err
	}

	fstabOutput, err := h.cmdExec.Execute(ctx, "genfstab", "-U", cmd.MountPoint)
	if err != nil {
		h.logger.Error("Failed to generate fstab", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to generate fstab: %v", err)
		return result, err
	}

	fstabPath := filepath.Join(cmd.MountPoint, "etc", "fstab")
	if err := h.fs.WriteFile(fstabPath, fstabOutput, 0644); err != nil {
		h.logger.Error("Failed to write fstab", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to write fstab: %v", err)
		return result, err
	}

	result.PackagesInstalled = basePackages
	result.Success = true

	h.logger.Info("Base system installation completed", "packages", len(basePackages))
	return result, nil
}

// setupCachyOSOnHost configures the CachyOS repo on the live host so that
// pacstrap can resolve linux-cachyos. Mirrors the working bash approach:
// import+sign key, write mirrorlist manually (no versioned package URLs),
// patch pacman.conf, then sync.
func (h *InstallBaseHandler) setupCachyOSOnHost(ctx context.Context) error {
	h.logger.Info("Setting up CachyOS repo on host for pacstrap")

	const (
		cachyOSKeyID          = "F3B607488DB35A47"
		keyserver             = "keyserver.ubuntu.com"
		hostPacmanConf        = "/etc/pacman.conf"
		cachyOSMirrorlistPath = "/etc/pacman.d/cachyos-mirrorlist"
		cachyOSMirrorlist     = "## CachyOS mirrorlist\nServer = https://mirror.cachyos.org/repo/$arch/$repo\n"
	)

	// Only fetch from keyserver if key is not already present.
	if _, err := h.cmdExec.Execute(ctx, "pacman-key", "--list-keys", cachyOSKeyID); err != nil {
		if _, err := h.cmdExec.Execute(ctx, "pacman-key", "--recv-keys",
			cachyOSKeyID, "--keyserver", keyserver); err != nil {
			return fmt.Errorf("failed to receive CachyOS key on host: %w", err)
		}
	}

	// Local-sign the key so pacman trusts packages from this unofficial repo.
	if _, err := h.cmdExec.Execute(ctx, "pacman-key", "--lsign-key", cachyOSKeyID); err != nil {
		return fmt.Errorf("failed to sign CachyOS key on host: %w", err)
	}

	// Write static mirrorlist — no versioned package URLs that can go stale.
	if err := h.fs.WriteFile(cachyOSMirrorlistPath, []byte(cachyOSMirrorlist), 0644); err != nil {
		return fmt.Errorf("failed to write host CachyOS mirrorlist: %w", err)
	}

	confBytes, err := h.fs.ReadFile(hostPacmanConf)
	if err != nil {
		return fmt.Errorf("failed to read host pacman.conf: %w", err)
	}
	conf := ensureCachyOSHostRepo(string(confBytes))
	if err := h.fs.WriteFile(hostPacmanConf, []byte(conf), 0644); err != nil {
		return fmt.Errorf("failed to write host pacman.conf: %w", err)
	}

	if _, err := h.cmdExec.Execute(ctx, "pacman", "-Sy", "--noconfirm"); err != nil {
		return fmt.Errorf("failed to sync host pacman repos: %w", err)
	}

	h.logger.Info("CachyOS repo ready on host")
	return nil
}

func ensureCachyOSHostRepo(conf string) string {
	if strings.Contains(conf, "[cachyos]") {
		return conf
	}

	conf = strings.TrimRight(conf, "\n")
	if conf == "" {
		return "[cachyos]\nInclude = /etc/pacman.d/cachyos-mirrorlist\n"
	}

	block := []string{
		"# CachyOS repositories",
		"[cachyos]",
		"Include = /etc/pacman.d/cachyos-mirrorlist",
		"",
	}

	lines := strings.Split(conf, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "[core]" {
			out := make([]string, 0, len(lines)+len(block))
			out = append(out, lines[:i]...)
			out = append(out, block...)
			out = append(out, lines[i:]...)
			return strings.Join(out, "\n") + "\n"
		}
	}

	// Fallback: append if [core] section is not found.
	return conf + "\n[cachyos]\nInclude = /etc/pacman.d/cachyos-mirrorlist\n"
}

// loadBasePackages reads the base package list from file
func (h *InstallBaseHandler) loadBasePackages() ([]string, error) {
	// Read from install directory (downloaded during bootstrap)
	packageFile := filepath.Join(config.DefaultInstallDir, config.BasePackagesFile)

	content, err := h.fs.ReadFile(packageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", packageFile, err)
	}

	var packages []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		packages = append(packages, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading package file: %w", err)
	}

	h.logger.Info("Loaded packages from file", "file", packageFile, "count", len(packages))
	return packages, nil
}
