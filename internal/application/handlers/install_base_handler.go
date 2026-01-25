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

	// Add any additional packages from command
	if len(cmd.Packages) > 0 {
		basePackages = append(basePackages, cmd.Packages...)
		h.logger.Info("Adding additional packages", "count", len(cmd.Packages))
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
