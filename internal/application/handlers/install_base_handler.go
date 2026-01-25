package handlers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
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

	// Build base packages list
	basePackages := []string{
		"base",
		"linux-firmware",
		kernel.PackageName(),
	}

	// Add microcode if requested
	if cmd.IncludeMicrocode {
		basePackages = append(basePackages, "intel-ucode", "amd-ucode")
		h.logger.Info("Adding CPU microcode packages")
	}

	// Add custom packages
	basePackages = append(basePackages, cmd.Packages...)

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
