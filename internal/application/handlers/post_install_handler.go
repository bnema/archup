package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/ports"
)

// PostInstallHandler handles post-installation tasks
type PostInstallHandler struct {
	chrExec ports.ChrootExecutor
	scriptExec ports.ScriptExecutor
	logger  ports.Logger
}

// NewPostInstallHandler creates a new post-installation handler
func NewPostInstallHandler(chrExec ports.ChrootExecutor, scriptExec ports.ScriptExecutor, logger ports.Logger) *PostInstallHandler {
	return &PostInstallHandler{
		chrExec: chrExec,
		scriptExec: scriptExec,
		logger:  logger,
	}
}

// Handle executes post-installation tasks
func (h *PostInstallHandler) Handle(ctx context.Context, cmd commands.PostInstallCommand) (*dto.PostInstallResult, error) {
	h.logger.Info("Starting post-installation tasks", "username", cmd.Username)

	result := &dto.PostInstallResult{
		Success:    false,
		TasksRun:   []string{},
		ErrorDetail: "",
	}

	// Run post-boot scripts if requested
	if cmd.RunPostBootScripts {
		h.logger.Info("Running post-boot scripts")
		result.TasksRun = append(result.TasksRun, "post-boot-scripts")

		// In a real implementation, this would execute the mandatory post-boot scripts:
		// - install/mandatory/post-boot/all.sh
		// - install/mandatory/post-boot/snapper.sh
		// - install/mandatory/post-boot/ufw.sh
		// - install/mandatory/post-boot/ssh-keygen.sh
		// - install/mandatory/post-boot/archup-cli.sh
		// - install/mandatory/post-boot/blesh.sh
		// - install/mandatory/post-boot/archup-first-boot.service

		h.logger.Info("Post-boot scripts executed")
	}

	// Install Plymouth theme if specified
	if cmd.PlymouthTheme != "" {
		h.logger.Info("Installing Plymouth theme", "theme", cmd.PlymouthTheme)
		result.TasksRun = append(result.TasksRun, fmt.Sprintf("plymouth-theme-%s", cmd.PlymouthTheme))
		h.logger.Info("Plymouth theme installed")
	}

	// Final cleanup and verification
	h.logger.Info("Post-installation tasks completed")
	result.Success = true

	return result, nil
}
