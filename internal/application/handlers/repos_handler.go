package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/packages"
	"github.com/bnema/archup/internal/domain/ports"
)

// ReposHandler handles repository configuration
type ReposHandler struct {
	fs      ports.FileSystem
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewReposHandler creates a new repositories handler
func NewReposHandler(fs ports.FileSystem, chrExec ports.ChrootExecutor, logger ports.Logger) *ReposHandler {
	return &ReposHandler{
		fs:      fs,
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

	fail := func(message string, err error) (*dto.RepositoriesResult, error) {
		h.logger.Error(message, "error", err)
		result.ErrorDetail = fmt.Sprintf("%s: %v", message, err)
		return result, err
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

	pacmanConfPath := filepath.Join(cmd.MountPoint, "etc", "pacman.conf")
	if repo.EnableMultilib() || repo.EnableChaotic() {
		confBytes, err := h.fs.ReadFile(pacmanConfPath)
		if err != nil {
			return fail("Failed to read pacman.conf", err)
		}

		conf := string(confBytes)
		if repo.EnableMultilib() {
			conf = enableMultilibSection(conf)
		}
		if repo.EnableChaotic() {
			conf = ensureChaoticRepo(conf)
		}

		if err := h.fs.WriteFile(pacmanConfPath, []byte(conf), 0644); err != nil {
			return fail("Failed to write pacman.conf", err)
		}
	}

	if repo.EnableChaotic() {
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman-key", "--recv-key", "3056513887B78AEB", "--keyserver", "keyserver.ubuntu.com"); err != nil {
			return fail("Failed to receive Chaotic-AUR key", err)
		}

		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman-key", "--lsign-key", "3056513887B78AEB"); err != nil {
			return fail("Failed to sign Chaotic-AUR key", err)
		}

		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman", "-U", "--noconfirm",
			"https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst",
			"https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst"); err != nil {
			return fail("Failed to install Chaotic-AUR keyring", err)
		}
	}

	if repo.EnableMultilib() || repo.EnableChaotic() {
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman", "-Sy", "--noconfirm"); err != nil {
			return fail("Failed to sync pacman repositories", err)
		}
	}

	if repo.EnableChaotic() {
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman", "-S", "--noconfirm", repo.AURHelper().String()); err != nil {
			return fail("Failed to install AUR helper", err)
		}
	}

	result.Success = true
	result.AURHelper = repo.AURHelper().String()

	h.logger.Info("Repository configuration completed successfully")
	return result, nil
}

func enableMultilibSection(conf string) string {
	lines := strings.Split(conf, "\n")
	inSection := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#[multilib]"):
			lines[i] = strings.Replace(line, "#", "", 1)
			inSection = true
		case strings.HasPrefix(trimmed, "[multilib]"):
			inSection = true
		case strings.HasPrefix(trimmed, "["):
			inSection = false
		case inSection && strings.HasPrefix(trimmed, "#Include"):
			lines[i] = strings.Replace(line, "#", "", 1)
		}
	}
	return strings.Join(lines, "\n")
}

func ensureChaoticRepo(conf string) string {
	if strings.Contains(conf, "[chaotic-aur]") {
		return conf
	}

	conf = strings.TrimRight(conf, "\n")
	chaotic := "\n[chaotic-aur]\nInclude = /etc/pacman.d/chaotic-mirrorlist\n"
	return conf + chaotic
}
