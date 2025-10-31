package handlers

import (
	"context"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/ports"
)

// PreflightHandler handles preflight checks
type PreflightHandler struct {
	fs       ports.FileSystem
	cmdExec  ports.CommandExecutor
	logger   ports.Logger
}

// NewPreflightHandler creates a new preflight handler
func NewPreflightHandler(fs ports.FileSystem, cmdExec ports.CommandExecutor, logger ports.Logger) *PreflightHandler {
	return &PreflightHandler{
		fs:      fs,
		cmdExec: cmdExec,
		logger:  logger,
	}
}

// Handle executes preflight checks
func (h *PreflightHandler) Handle(ctx context.Context, cmd commands.PreflightCommand) (*dto.PreflightResult, error) {
	h.logger.Info("Starting preflight checks")

	result := &dto.PreflightResult{
		SystemInfo:     &dto.SystemInfo{},
		CPUInfo:        &dto.CPUInfo{},
		ChecksPassed:   true,
		Warnings:       []string{},
		CriticalErrors: []string{},
	}

	// Check if running as root
	if output, err := h.cmdExec.Execute(ctx, "id", "-u"); err == nil {
		if string(output) != "0\n" {
			result.CriticalErrors = append(result.CriticalErrors, "Must run as root user")
			result.ChecksPassed = false
			h.logger.Warn("Not running as root")
		}
	}

	// Check architecture
	if output, err := h.cmdExec.Execute(ctx, "uname", "-m"); err == nil && len(output) > 0 {
		arch := string(output)
		if arch[len(arch)-1] == '\n' {
			arch = arch[:len(arch)-1]
		}
		result.SystemInfo.Architecture = arch
		h.logger.Info("Detected architecture", "arch", result.SystemInfo.Architecture)
	}

	// Check UEFI boot
	exists, _ := h.fs.Exists("/sys/firmware/efi/fw_platform_size")
	result.SystemInfo.IsUEFI = exists
	if !exists {
		result.Warnings = append(result.Warnings, "Not booted in UEFI mode")
	}

	// Check internet connectivity (try DNS)
	if _, err := h.cmdExec.Execute(ctx, "ping", "-c", "1", "archlinux.org"); err != nil {
		result.Warnings = append(result.Warnings, "Could not verify internet connectivity")
	}

	// Get CPU info
	if output, err := h.cmdExec.Execute(ctx, "grep", "model name", "/proc/cpuinfo"); err == nil && len(output) > 0 {
		model := string(output)
		if model[len(model)-1] == '\n' {
			model = model[:len(model)-1]
		}
		result.CPUInfo.Model = model
	}

	if result.ChecksPassed {
		h.logger.Info("Preflight checks passed")
	} else {
		h.logger.Error("Preflight checks failed", "errors", result.CriticalErrors)
	}

	return result, nil
}
