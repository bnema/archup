package handlers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
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

	// Write X11/Wayland keyboard layout so GUI sessions use the same keymap as the console.
	if err := h.writeX11KeyboardConf(cmd.MountPoint, sysConfig.Keymap()); err != nil {
		// Non-fatal: log and continue. Console keymap is already set via vconsole.conf.
		h.logger.Warn("Failed to write X11 keyboard config", "error", err)
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

	// Mask NetworkManager-wait-online to prevent boot delays when no network is immediately available.
	if err := h.chrExec.ChrootSystemctl(ctx, h.logger.LogPath(), cmd.MountPoint, "mask", "NetworkManager-wait-online.service"); err != nil {
		h.logger.Warn("Failed to mask NetworkManager-wait-online.service", "error", err)
	}

	// Migrate iwd WiFi credentials from live ISO to NetworkManager profiles in the installed system.
	if err := h.migrateIWDProfiles(cmd.MountPoint); err != nil {
		// Non-fatal: log and continue. User can configure WiFi manually after boot.
		h.logger.Warn("Failed to migrate iwd WiFi profiles", "error", err)
	}

	h.logger.Info("System configuration completed successfully")

	result.Success = true
	result.Timezone = sysConfig.Timezone()
	result.Username = usr.Username()

	return result, nil
}

// writeX11KeyboardConf writes /etc/X11/xorg.conf.d/00-keyboard.conf so that
// X11 and Wayland sessions use the same keyboard layout as the console.
// The XKB layout is derived from the keymap by stripping any variant suffix
// (e.g. "de-nodeadkeys" → "de", "fr" → "fr").
func (h *ConfigureSystemHandler) writeX11KeyboardConf(mountPoint, keymap string) error {
	xkbLayout := xkbLayoutFromKeymap(keymap)
	if xkbLayout == "" {
		return nil
	}

	xorgDir := filepath.Join(mountPoint, "etc", "X11", "xorg.conf.d")
	if err := h.fs.MkdirAll(xorgDir, 0755); err != nil {
		return fmt.Errorf("failed to create xorg.conf.d directory: %w", err)
	}

	content := fmt.Sprintf(`Section "InputClass"
        Identifier "system-keyboard"
        MatchIsKeyboard "on"
        Option "XkbLayout" "%s"
EndSection
`, xkbLayout)

	confPath := filepath.Join(xorgDir, "00-keyboard.conf")
	if err := h.fs.WriteFile(confPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write 00-keyboard.conf: %w", err)
	}

	return nil
}

// xkbLayoutFromKeymap extracts the XKB layout from a vconsole keymap name.
// Keymap variants use hyphens (e.g. "de-nodeadkeys"), XKB uses the base part.
func xkbLayoutFromKeymap(keymap string) string {
	if keymap == "" {
		return ""
	}
	// Strip variant suffix: "de-nodeadkeys" → "de", "fr" → "fr"
	if idx := strings.Index(keymap, "-"); idx != -1 {
		return keymap[:idx]
	}
	return keymap
}

// migrateIWDProfiles reads *.psk files from the live ISO's /var/lib/iwd and
// writes equivalent NetworkManager keyfile profiles into the installed system.
// This ensures the system has WiFi connectivity on first boot without manual setup.
func (h *ConfigureSystemHandler) migrateIWDProfiles(mountPoint string) error {
	const iwdDir = "/var/lib/iwd"

	entries, err := os.ReadDir(iwdDir)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Info("No iwd profiles found (probably ethernet install), skipping WiFi migration")
			return nil
		}
		return fmt.Errorf("failed to read iwd directory: %w", err)
	}

	nmConnDir := filepath.Join(mountPoint, "etc", "NetworkManager", "system-connections")
	if err := h.fs.MkdirAll(nmConnDir, 0700); err != nil {
		return fmt.Errorf("failed to create NM connections directory: %w", err)
	}

	migrated := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".psk") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(iwdDir, entry.Name()))
		if err != nil {
			h.logger.Warn("Failed to read iwd profile", "file", entry.Name(), "error", err)
			continue
		}

		// SSID is the filename without the .psk extension.
		ssid := strings.TrimSuffix(entry.Name(), ".psk")
		// iwd hex-encodes SSIDs containing non-alphanumeric chars as =<hex>.
		if strings.HasPrefix(ssid, "=") {
			decoded, err := decodeIWDHexSSID(ssid[1:])
			if err == nil {
				ssid = decoded
			}
		}

		psk := extractIWDValue(data, "PreSharedKey")
		passphrase := extractIWDValue(data, "Passphrase")

		// Prefer the raw passphrase for NM; fall back to PreSharedKey (hex PSK).
		password := passphrase
		if password == "" {
			password = psk
		}
		if password == "" {
			h.logger.Warn("iwd profile has no usable password, skipping", "ssid", ssid)
			continue
		}

		profile := buildNMWiFiProfile(ssid, password)
		profilePath := filepath.Join(nmConnDir, ssid+".nmconnection")
		if err := h.fs.WriteFile(profilePath, []byte(profile), 0600); err != nil {
			h.logger.Warn("Failed to write NM WiFi profile", "ssid", ssid, "error", err)
			continue
		}

		h.logger.Info("Migrated WiFi profile", "ssid", ssid)
		migrated++
	}

	if migrated > 0 {
		h.logger.Info("WiFi profiles migrated to NetworkManager", "count", migrated)
	}
	return nil
}

// extractIWDValue parses a key=value line from an iwd .psk file.
func extractIWDValue(data []byte, key string) string {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	prefix := key + "="
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix)
		}
	}
	return ""
}

// decodeIWDHexSSID decodes an iwd hex-encoded SSID (used when SSID contains
// non-alphanumeric characters). Format: lowercase hex pairs.
func decodeIWDHexSSID(hex string) (string, error) {
	if len(hex)%2 != 0 {
		return "", fmt.Errorf("odd hex length")
	}
	b := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		hi := hexNibble(hex[i])
		lo := hexNibble(hex[i+1])
		if hi < 0 || lo < 0 {
			return "", fmt.Errorf("invalid hex char")
		}
		b[i/2] = byte(hi<<4 | lo)
	}
	return string(b), nil
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

// buildNMWiFiProfile generates a NetworkManager keyfile for a WPA-PSK network.
func buildNMWiFiProfile(ssid, psk string) string {
	return fmt.Sprintf(`[connection]
id=%s
type=wifi
autoconnect=true

[wifi]
ssid=%s
mode=infrastructure

[wifi-security]
auth-alg=open
key-mgmt=wpa-psk
psk=%s

[ipv4]
method=auto

[ipv6]
method=auto
`, ssid, ssid, psk)
}
