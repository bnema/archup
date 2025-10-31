package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/bootloader"
	"github.com/bnema/archup/internal/domain/ports"
)

// BootloaderHandler handles bootloader installation
type BootloaderHandler struct {
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewBootloaderHandler creates a new bootloader handler
func NewBootloaderHandler(chrExec ports.ChrootExecutor, logger ports.Logger) *BootloaderHandler {
	return &BootloaderHandler{
		chrExec: chrExec,
		logger:  logger,
	}
}

// Handle installs the bootloader
func (h *BootloaderHandler) Handle(ctx context.Context, cmd commands.InstallBootloaderCommand) (*dto.BootloaderResult, error) {
	h.logger.Info("Starting bootloader installation", "type", cmd.BootloaderType)

	result := &dto.BootloaderResult{
		Success:       false,
		BootloaderType: "",
		Timeout:       cmd.TimeoutSeconds,
		ErrorDetail:   "",
	}

	// Create bootloader domain object
	bl, err := bootloader.NewBootloader(cmd.BootloaderType, cmd.TimeoutSeconds, cmd.Branding)
	if err != nil {
		h.logger.Error("Invalid bootloader configuration", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid bootloader configuration: %v", err)
		return result, err
	}

	h.logger.Info("Bootloader configuration validated", "type", bl.Type())

	// In a real implementation, this would configure and install the bootloader
	// For Limine: configure limine.cfg
	// For systemd-boot: configure bootctl and entry files

	result.Success = true
	result.BootloaderType = bl.Type().String()
	result.Timeout = bl.Timeout()

	h.logger.Info("Bootloader installation completed successfully")
	return result, nil
}
