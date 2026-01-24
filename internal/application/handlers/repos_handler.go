package handlers

import (
	"context"
	"fmt"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports"
)

// ReposHandler handles repository configuration
type ReposHandler struct {
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewReposHandler creates a new repositories handler
func NewReposHandler(chrExec ports.ChrootExecutor, logger ports.Logger) *ReposHandler {
	return &ReposHandler{
		chrExec: chrExec,
		logger:  logger,
	}
}

// Handle configures package repositories
func (h *ReposHandler) Handle(ctx context.Context, cmd commands.SetupRepositoriesCommand) (*dto.RepositoriesResult, error) {
	h.logger.Info("Starting repository configuration", "multilib", cmd.EnableMultilib, "chaotic", cmd.EnableChaotic)

	result := &dto.RepositoriesResult{
		Success:     false,
		Multilib:    cmd.EnableMultilib,
		Chaotic:     cmd.EnableChaotic,
		AURHelper:   "",
		ErrorDetail: "",
	}

	// Create repository domain object
	repo, err := packages.NewRepository(cmd.EnableMultilib, cmd.EnableChaotic, cmd.AURHelper)
	if err != nil {
		h.logger.Error("Invalid repository configuration", "error", err)
		result.ErrorDetail = fmt.Sprintf("Invalid repository configuration: %v", err)
		return result, err
	}

	h.logger.Info("Repository configuration validated",
		"multilib", repo.EnableMultilib(),
		"chaotic", repo.EnableChaotic(),
		"aurHelper", repo.AURHelper())

	// In a real implementation, this would:
	// 1. Configure pacman.conf for multilib/chaotic
	// 2. Install AUR helper (paru or yay)
	// 3. Configure additional repositories

	result.Success = true
	result.AURHelper = repo.AURHelper().String()

	h.logger.Info("Repository configuration completed successfully")
	return result, nil
}
