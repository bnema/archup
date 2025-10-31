package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/system"
	"github.com/bnema/archup/internal/domain/user"
)

// ConfigureSystemHandler handles system configuration
type ConfigureSystemHandler struct {
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewConfigureSystemHandler creates a new system configuration handler
func NewConfigureSystemHandler(chrExec ports.ChrootExecutor, logger ports.Logger) *ConfigureSystemHandler {
	return &ConfigureSystemHandler{
		chrExec: chrExec,
		logger:  logger,
	}
}

// Handle configures the system
func (h *ConfigureSystemHandler) Handle(ctx context.Context, cmd commands.ConfigureSystemCommand) (*dto.ConfigureSystemResult, error) {
	h.logger.Info("Starting system configuration", "hostname", cmd.Hostname)

	result := &dto.ConfigureSystemResult{
		Success:     false,
		Hostname:    cmd.Hostname,
		Timezone:    cmd.Timezone,
		ErrorDetail: "",
	}

	// Create system config domain object
	sysConfig, err := system.NewSystemConfig(cmd.Hostname, cmd.Timezone, cmd.Locale, cmd.Keymap)
	if err != nil {
		h.logger.Error("Invalid system configuration", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid system configuration: %v", err)
		return result, err
	}

	h.logger.Info("System configuration validated", "hostname", sysConfig.Hostname())

	// Create user domain object
	usr, err := user.NewUser(cmd.Username, cmd.UserShell)
	if err != nil {
		h.logger.Error("Invalid user configuration", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid user: %v", err)
		return result, err
	}

	// Add user to wheel group for sudo
	if err := usr.AddGroup("wheel"); err != nil {
		h.logger.Warn("Failed to add user to wheel group", "error", err)
	}

	// Create credentials (validates password strength and constraints)
	_, err = user.NewCredentials(cmd.UserPassword, cmd.RootPassword)
	if err != nil {
		h.logger.Error("Invalid credentials", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid credentials: %v", err)
		return result, err
	}

	h.logger.Info("User configuration validated", "username", usr.Username())
	h.logger.Info("Credentials validated successfully")
	h.logger.Info("System configuration completed successfully")

	result.Success = true
	result.Timezone = sysConfig.Timezone()
	result.Username = usr.Username()

	return result, nil
}
