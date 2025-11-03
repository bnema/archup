package handlers

import (
	"context"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/system"
)

// PreflightHandler handles preflight checks
type PreflightHandler struct {
	fs       ports.FileSystem
	cmdExec  ports.CommandExecutor
	logger   ports.Logger
	rules    *system.SystemValidationRules
}

// NewPreflightHandler creates a new preflight handler
func NewPreflightHandler(fs ports.FileSystem, cmdExec ports.CommandExecutor, logger ports.Logger) *PreflightHandler {
	return &PreflightHandler{
		fs:      fs,
		cmdExec: cmdExec,
		logger:  logger,
		rules:   system.NewSystemValidationRules(),
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
		if strings.TrimSpace(string(output)) != "0" {
			result.CriticalErrors = append(result.CriticalErrors, "Must run as root user")
			result.ChecksPassed = false
			h.logger.Warn("Not running as root")
		}
	}

	// Check for Arch Linux
	if err := h.rules.ValidateArchLinux(ctx, h.fs); err != nil {
		result.CriticalErrors = append(result.CriticalErrors, err.Error())
		result.ChecksPassed = false
		h.logger.Error("Not running on Arch Linux", "error", err)
	}

	// Detect distribution type
	if distro, err := system.DetectDistribution(ctx, h.fs); err == nil {
		result.SystemInfo.Distribution = distro.String()
		h.logger.Info("Detected distribution", "distro", distro.String())

		// Check for derivatives (must be vanilla Arch)
		if err := h.rules.ValidateNotDerivative(ctx, h.fs); err != nil {
			result.CriticalErrors = append(result.CriticalErrors, err.Error())
			result.ChecksPassed = false
			h.logger.Error("Derivative distribution detected", "error", err)
		}
	}

	// Check architecture
	if output, err := h.cmdExec.Execute(ctx, "uname", "-m"); err == nil && len(output) > 0 {
		arch := strings.TrimSpace(string(output))
		result.SystemInfo.Architecture = arch
		h.logger.Info("Detected architecture", "arch", arch)

		// Validate x86_64
		if err := h.rules.ValidateArchitecture(arch); err != nil {
			result.CriticalErrors = append(result.CriticalErrors, err.Error())
			result.ChecksPassed = false
			h.logger.Error("Invalid architecture", "error", err)
		}
	}

	// Check UEFI boot
	if err := h.rules.ValidateUEFIBoot(ctx, h.fs); err != nil {
		result.CriticalErrors = append(result.CriticalErrors, err.Error())
		result.ChecksPassed = false
		result.SystemInfo.IsUEFI = false
		h.logger.Error("Not booted in UEFI mode", "error", err)
	} else {
		result.SystemInfo.IsUEFI = true
		h.logger.Info("UEFI boot mode confirmed")
	}

	// Check Secure Boot status
	if enabled, err := h.rules.DetectSecureBoot(ctx, h.cmdExec); err == nil {
		result.SystemInfo.SecureBootEnabled = enabled
		if enabled {
			h.logger.Warn("Secure Boot is enabled")
		} else {
			h.logger.Info("Secure Boot is disabled")
		}

		// Validate that Secure Boot is disabled
		if err := h.rules.ValidateSecureBootDisabled(ctx, h.cmdExec); err != nil {
			result.CriticalErrors = append(result.CriticalErrors, err.Error())
			result.ChecksPassed = false
			h.logger.Error("Secure Boot must be disabled", "error", err)
		}
	} else {
		h.logger.Warn("Could not detect Secure Boot status", "error", err)
	}

	// Check internet connectivity (try DNS)
	if _, err := h.cmdExec.Execute(ctx, "ping", "-c", "1", "archlinux.org"); err != nil {
		result.Warnings = append(result.Warnings, "Could not verify internet connectivity")
		h.logger.Warn("Internet connectivity check failed")
	}

	// Get CPU info
	if output, err := h.cmdExec.Execute(ctx, "grep", "model name", "/proc/cpuinfo"); err == nil && len(output) > 0 {
		model := strings.TrimSpace(string(output))
		result.CPUInfo.Model = model
		h.logger.Info("Detected CPU", "model", model)
	}

	if result.ChecksPassed {
		h.logger.Info("Preflight checks passed")
	} else {
		h.logger.Error("Preflight checks failed", "errors", result.CriticalErrors)
	}

	return result, nil
}
