package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/application/commands"
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/system"
	"github.com/bnema/archup/internal/domain/user"
)

// ConfigureSystemHandler handles system configuration
type ConfigureSystemHandler struct {
	fs      ports.FileSystem
	chrExec ports.ChrootExecutor
	logger  ports.Logger
}

// NewConfigureSystemHandler creates a new system configuration handler
func NewConfigureSystemHandler(fs ports.FileSystem, chrExec ports.ChrootExecutor, logger ports.Logger) *ConfigureSystemHandler {
	return &ConfigureSystemHandler{
		fs:      fs,
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

	if err := h.fs.WriteFile(filepath.Join(cmd.MountPoint, "etc", "hostname"), []byte(sysConfig.Hostname()+"\n"), 0644); err != nil {
		h.logger.Error("Failed to write hostname", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to write hostname: %v", err)
		return result, err
	}

	hostsContent := fmt.Sprintf("127.0.0.1\tlocalhost\n::1\tlocalhost\n127.0.1.1\t%s.localdomain\t%s\n", sysConfig.Hostname(), sysConfig.Hostname())
	if err := h.fs.WriteFile(filepath.Join(cmd.MountPoint, "etc", "hosts"), []byte(hostsContent), 0644); err != nil {
		h.logger.Error("Failed to write hosts file", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to write hosts file: %v", err)
		return result, err
	}

	if sysConfig.Timezone() != "" {
		zonePath := filepath.Join("/usr/share/zoneinfo", sysConfig.Timezone())
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "ln", "-sf", zonePath, "/etc/localtime"); err != nil {
			h.logger.Error("Failed to set timezone", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to set timezone: %v", err)
			return result, err
		}
	}

	if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "hwclock", "--systohc"); err != nil {
		h.logger.Error("Failed to set hardware clock", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to set hardware clock: %v", err)
		return result, err
	}

	if sysConfig.Locale() != "" {
		localeGen := fmt.Sprintf("%s UTF-8\n", sysConfig.Locale())
		if err := h.fs.WriteFile(filepath.Join(cmd.MountPoint, "etc", "locale.gen"), []byte(localeGen), 0644); err != nil {
			h.logger.Error("Failed to write locale.gen", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to write locale.gen: %v", err)
			return result, err
		}

		localeConf := fmt.Sprintf("LANG=%s\n", sysConfig.Locale())
		if err := h.fs.WriteFile(filepath.Join(cmd.MountPoint, "etc", "locale.conf"), []byte(localeConf), 0644); err != nil {
			h.logger.Error("Failed to write locale.conf", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to write locale.conf: %v", err)
			return result, err
		}

		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "locale-gen"); err != nil {
			h.logger.Error("Failed to generate locale", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to generate locale: %v", err)
			return result, err
		}
	}

	if err := h.fs.WriteFile(filepath.Join(cmd.MountPoint, "etc", "vconsole.conf"), []byte("KEYMAP="+sysConfig.Keymap()+"\n"), 0644); err != nil {
		h.logger.Error("Failed to write vconsole.conf", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to write vconsole.conf: %v", err)
		return result, err
	}

	groups := strings.Join(usr.Groups(), ",")
	if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "useradd", "-m", "-G", groups, "-s", usr.Shell(), usr.Username()); err != nil {
		h.logger.Error("Failed to create user", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create user: %v", err)
		return result, err
	}

	userCreds := fmt.Sprintf("%s:%s\n", usr.Username(), cmd.UserPassword)
	if err := h.chrExec.ExecuteInChrootWithStdin(ctx, cmd.MountPoint, userCreds, "chpasswd"); err != nil {
		h.logger.Error("Failed to set user password", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to set user password: %v", err)
		return result, err
	}

	if cmd.RootPassword != "" {
		rootCreds := fmt.Sprintf("root:%s\n", cmd.RootPassword)
		if err := h.chrExec.ExecuteInChrootWithStdin(ctx, cmd.MountPoint, rootCreds, "chpasswd"); err != nil {
			h.logger.Error("Failed to set root password", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to set root password: %v", err)
			return result, err
		}
	} else {
		if _, err := h.chrExec.ExecuteInChroot(ctx, cmd.MountPoint, "passwd", "-l", "root"); err != nil {
			h.logger.Error("Failed to lock root account", "error", err)
			result.ErrorDetail = fmt.Sprintf("Failed to lock root account: %v", err)
			return result, err
		}
	}

	sudoersDir := filepath.Join(cmd.MountPoint, "etc", "sudoers.d")
	if err := h.fs.MkdirAll(sudoersDir, 0755); err != nil {
		h.logger.Error("Failed to create sudoers.d directory", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to create sudoers.d directory: %v", err)
		return result, err
	}
	if err := h.fs.WriteFile(filepath.Join(sudoersDir, "wheel"), []byte("%wheel ALL=(ALL:ALL) ALL\n"), 0440); err != nil {
		h.logger.Error("Failed to write sudoers", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to write sudoers: %v", err)
		return result, err
	}

	if err := h.chrExec.ChrootSystemctl(ctx, h.logger.LogPath(), cmd.MountPoint, "enable", "NetworkManager"); err != nil {
		h.logger.Error("Failed to enable NetworkManager", "error", err)
		result.ErrorDetail = fmt.Sprintf("Failed to enable NetworkManager: %v", err)
		return result, err
	}

	h.logger.Info("System configuration completed successfully")

	result.Success = true
	result.Timezone = sysConfig.Timezone()
	result.Username = usr.Username()

	return result, nil
}
