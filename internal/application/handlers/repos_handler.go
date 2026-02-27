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

	// Enable multilib first (doesn't require external files)
	if repo.EnableMultilib() {
		confBytes, err := h.fs.ReadFile(pacmanConfPath)
		if err != nil {
			return fail("Failed to read pacman.conf", err)
		}
		conf := enableMultilibSection(string(confBytes))
		if err := h.fs.WriteFile(pacmanConfPath, []byte(conf), 0644); err != nil {
			return fail("Failed to write pacman.conf", err)
		}
	}

	// For CachyOS kernel: setup CachyOS repo BEFORE sync
	if cmd.KernelVariant == packages.KernelCachyOS {
		if err := h.setupCachyOSRepo(ctx, cmd.MountPoint); err != nil {
			return fail("Failed to setup CachyOS repository", err)
		}
	}

	// For chaotic-aur: install keyring and mirrorlist BEFORE adding to pacman.conf
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

		// Now that mirrorlist exists, add chaotic-aur to pacman.conf
		confBytes, err := h.fs.ReadFile(pacmanConfPath)
		if err != nil {
			return fail("Failed to read pacman.conf", err)
		}
		conf := ensureChaoticRepo(string(confBytes))
		if err := h.fs.WriteFile(pacmanConfPath, []byte(conf), 0644); err != nil {
			return fail("Failed to write pacman.conf", err)
		}
	}

	// Sync repos after all config changes
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

	// Install extra packages from extra.packages file
	extraPkgs, err := h.loadExtraPackages()
	if err != nil {
		h.logger.Warn("Could not load extra.packages", "error", err)
	} else if len(extraPkgs) > 0 {
		h.logger.Info("Installing extra packages", "count", len(extraPkgs))
		args := append([]string{"-S", "--noconfirm", "--needed"}, extraPkgs...)
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "pacman", args...); err != nil {
			return fail("Failed to install extra packages", err)
		}
		h.logger.Info("Extra packages installed successfully")
	}

	result.Success = true
	result.AURHelper = repo.AURHelper().String()

	h.logger.Info("Repository configuration completed successfully")
	return result, nil
}

// loadExtraPackages reads the extra package list from file
func (h *ReposHandler) loadExtraPackages() ([]string, error) {
	packageFile := filepath.Join(config.DefaultInstallDir, config.ExtraPackagesFile)

	content, err := h.fs.ReadFile(packageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", packageFile, err)
	}

	var pkgs []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pkgs = append(pkgs, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading package file: %w", err)
	}

	h.logger.Info("Loaded extra packages from file", "file", packageFile, "count", len(pkgs))
	return pkgs, nil
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

// setupCachyOSRepo configures the CachyOS repo inside the chroot. Mirrors the
// bash approach: import+sign key, write mirrorlist manually (no versioned
// package URLs), patch pacman.conf, then sync.
func (h *ReposHandler) setupCachyOSRepo(ctx context.Context, mountPoint string) error {
	const (
		cachyOSKeyID      = "F3B607488DB35A47"
		keyserver         = "keyserver.ubuntu.com"
		cachyOSMirrorlist = "## CachyOS mirrorlist\nServer = https://mirror.cachyos.org/repo/$arch/$repo\n"
	)

	// Only fetch from keyserver if key is not already present in chroot.
	if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "pacman-key", "--list-keys", cachyOSKeyID); err != nil {
		if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "pacman-key", "--recv-keys",
			cachyOSKeyID, "--keyserver", keyserver); err != nil {
			return fmt.Errorf("failed to receive CachyOS key: %w", err)
		}
	}

	if _, err := h.chrExec.ExecuteInChroot(ctx, mountPoint, "pacman-key", "--lsign-key", cachyOSKeyID); err != nil {
		return fmt.Errorf("failed to sign CachyOS key: %w", err)
	}

	// Write static mirrorlist — no versioned package URLs that can go stale.
	mirrorlistPath := filepath.Join(mountPoint, "etc", "pacman.d", "cachyos-mirrorlist")
	if err := h.fs.WriteFile(mirrorlistPath, []byte(cachyOSMirrorlist), 0644); err != nil {
		return fmt.Errorf("failed to write CachyOS mirrorlist: %w", err)
	}

	pacmanConfPath := filepath.Join(mountPoint, "etc", "pacman.conf")
	confBytes, err := h.fs.ReadFile(pacmanConfPath)
	if err != nil {
		return fmt.Errorf("failed to read pacman.conf: %w", err)
	}
	conf := ensureCachyOSRepo(string(confBytes))
	return h.fs.WriteFile(pacmanConfPath, []byte(conf), 0644)
}

func ensureCachyOSRepo(conf string) string {
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
