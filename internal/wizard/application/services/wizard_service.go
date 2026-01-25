package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	osuser "os/user"
	"path/filepath"
	"strings"

	"github.com/bnema/archup/internal/defaults"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/wizard/domain"
)

// WizardService orchestrates the desktop setup wizard flow.
type WizardService struct {
	fs      ports.FileSystem
	cmdExec ports.CommandExecutor
	logger  ports.Logger
}

// NewWizardService creates a new wizard service.
func NewWizardService(fs ports.FileSystem, cmdExec ports.CommandExecutor, logger ports.Logger) *WizardService {
	return &WizardService{
		fs:      fs,
		cmdExec: cmdExec,
		logger:  logger,
	}
}

// Start begins the wizard run.
func (s *WizardService) Start(ctx context.Context, config domain.DesktopConfig) error {
	return s.Install(ctx, config)
}

// Install performs package installation and theme setup.
func (s *WizardService) Install(ctx context.Context, config domain.DesktopConfig) error {
	s.logger.Info("Wizard install", "compositor", config.Compositor)
	packages, err := s.buildPackageList(config)
	if err != nil {
		return err
	}
	if err := s.installPackages(ctx, packages); err != nil {
		return err
	}
	if err := s.ensureBleuTheme(ctx); err != nil {
		return err
	}
	return nil
}

// ApplyConfig writes compositor configuration.
func (s *WizardService) ApplyConfig(config domain.DesktopConfig) error {
	return s.ApplyConfigWithMonitors(config, nil)
}

// ApplyConfigWithMonitors writes compositor configuration and monitor layout.
func (s *WizardService) ApplyConfigWithMonitors(config domain.DesktopConfig, monitors []MonitorConfig) error {
	if err := s.applyCompositorConfig(config); err != nil {
		return err
	}
	if err := s.ensureHyprlockConfig(); err != nil {
		return err
	}
	if err := s.ensureHypridleConfig(); err != nil {
		return err
	}
	if len(monitors) > 0 {
		if err := s.writeMonitorConfig(config, monitors); err != nil {
			return err
		}
	}
	if config.EnableSDDM {
		if err := s.ensureSDDMTheme(); err != nil {
			return err
		}
		if err := s.ensureSDDMConfig(config); err != nil {
			return err
		}
		if err := s.enableSDDMService(); err != nil {
			return err
		}
	}
	if err := s.ensureWaybarConfig(); err != nil {
		return err
	}
	if err := s.applyGtkTheme(); err != nil {
		return err
	}
	if err := s.enableUserServices(); err != nil {
		return err
	}
	return nil
}

func (s *WizardService) buildPackageList(config domain.DesktopConfig) ([]string, error) {
	packages := []string{}

	core, err := defaults.ReadPackageList("desktop-core")
	if err != nil {
		return nil, err
	}
	packages = append(packages, core...)

	video, err := defaults.ReadPackageList("video-accel")
	if err != nil {
		return nil, err
	}
	packages = append(packages, video...)

	gtk, err := defaults.ReadPackageList("gtk")
	if err != nil {
		return nil, err
	}
	packages = append(packages, gtk...)

	if config.Compositor == domain.CompositorNiri {
		compositor, err := defaults.ReadPackageList("compositor-niri")
		if err != nil {
			return nil, err
		}
		packages = append(packages, compositor...)
	}

	if config.Compositor == domain.CompositorHyprland {
		compositor, err := defaults.ReadPackageList("compositor-hyprland")
		if err != nil {
			return nil, err
		}
		packages = append(packages, compositor...)
	}

	if config.EnableSDDM {
		sddm, err := defaults.ReadPackageList("sddm")
		if err != nil {
			return nil, err
		}
		packages = append(packages, sddm...)
	}

	if config.InstallCliphist {
		packages = append(packages, "cliphist")
	}

	return uniquePackages(packages), nil
}

func (s *WizardService) installPackages(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		s.logger.Info("No packages to install")
		return nil
	}

	args := append([]string{"-S", "--needed", "--noconfirm"}, packages...)
	s.logger.Info("Installing desktop packages", "count", len(packages))
	if _, err := s.cmdExec.Execute(ctx, "pacman", args...); err != nil {
		return fmt.Errorf("install packages: %w", err)
	}

	return nil
}

func (s *WizardService) applyCompositorConfig(config domain.DesktopConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	switch config.Compositor {
	case domain.CompositorHyprland:
		return s.applyHyprlandConfig(home, config)
	case domain.CompositorNiri:
		return s.applyNiriConfig(home, config)
	default:
		return nil
	}
}

func (s *WizardService) ensureBleuTheme(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	themeDir := filepath.Join(home, ".config", "themes", "bleu-theme")
	exists, err := s.fs.Exists(themeDir)
	if err != nil {
		return fmt.Errorf("check bleu theme directory: %w", err)
	}
	if !exists {
		if err := s.fs.MkdirAll(filepath.Dir(themeDir), 0o755); err != nil {
			return fmt.Errorf("create theme parent directory: %w", err)
		}
		if _, err := s.cmdExec.Execute(ctx, "git", "clone", "https://github.com/bnema/bleu-theme", themeDir); err != nil {
			return fmt.Errorf("clone bleu-theme: %w", err)
		}
	}

	wallpaperSource := filepath.Join(themeDir, "wallpapers", "wp-bleu-1.png")
	wallpaperTarget := filepath.Join(home, "Pictures", "Wallpapers", "wp-bleu-1.png")
	if err := s.fs.MkdirAll(filepath.Dir(wallpaperTarget), 0o755); err != nil {
		return fmt.Errorf("create wallpaper directory: %w", err)
	}
	data, err := s.fs.ReadFile(wallpaperSource)
	if err != nil {
		return fmt.Errorf("read wallpaper: %w", err)
	}
	if err := s.fs.WriteFile(wallpaperTarget, data, 0o644); err != nil {
		return fmt.Errorf("write wallpaper: %w", err)
	}

	return nil
}

func (s *WizardService) writeMonitorConfig(config domain.DesktopConfig, monitors []MonitorConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	switch config.Compositor {
	case domain.CompositorHyprland:
		return s.writeHyprlandMonitorConfig(home, monitors)
	case domain.CompositorNiri:
		return s.writeNiriMonitorConfig(home, monitors)
	default:
		return nil
	}
}

func (s *WizardService) writeHyprlandMonitorConfig(home string, monitors []MonitorConfig) error {
	configDir := filepath.Join(home, ".config", "hypr")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create hypr config dir: %w", err)
	}

	lines := []string{"# ArchUp monitors"}
	for _, monitor := range monitors {
		if !monitor.Enabled {
			lines = append(lines, fmt.Sprintf("monitor = %s,disable", monitor.Name))
			continue
		}
		mode := fmt.Sprintf("%dx%d@%.2f", monitor.Width, monitor.Height, monitor.Refresh)
		pos := fmt.Sprintf("%dx%d", monitor.PosX, monitor.PosY)
		scale := fmt.Sprintf("%.2f", monitor.Scale)
		lines = append(lines, fmt.Sprintf("monitor = %s,%s,%s,%s", monitor.Name, mode, pos, scale))
	}
	lines = append(lines, "")

	path := filepath.Join(configDir, "archup-monitors.conf")
	if err := s.writeIfChanged(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write hyprland monitor config: %w", err)
	}

	includeLine := "source = ~/.config/hypr/archup-monitors.conf"
	mainPath := filepath.Join(configDir, "hyprland.conf")
	contents, err := s.readFileIfExists(mainPath)
	if err != nil {
		return err
	}
	if !containsLine(contents, includeLine) {
		contents = appendLine(contents, includeLine)
		if err := s.fs.WriteFile(mainPath, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("write hyprland config: %w", err)
		}
	}

	return nil
}

func (s *WizardService) writeNiriMonitorConfig(home string, monitors []MonitorConfig) error {
	configDir := filepath.Join(home, ".config", "niri")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create niri config dir: %w", err)
	}

	lines := []string{"// ArchUp monitors"}
	for _, monitor := range monitors {
		if !monitor.Enabled {
			lines = append(lines, fmt.Sprintf("output \"%s\" {", monitor.Name))
			lines = append(lines, "    off")
			lines = append(lines, "}")
			lines = append(lines, "")
			continue
		}
		mode := fmt.Sprintf("%dx%d@%.2f", monitor.Width, monitor.Height, monitor.Refresh)
		lines = append(lines, fmt.Sprintf("output \"%s\" {", monitor.Name))
		lines = append(lines, fmt.Sprintf("    mode \"%s\"", mode))
		lines = append(lines, fmt.Sprintf("    scale %.2f", monitor.Scale))
		lines = append(lines, fmt.Sprintf("    position x=%d y=%d", monitor.PosX, monitor.PosY))
		lines = append(lines, "}")
		lines = append(lines, "")
	}

	path := filepath.Join(configDir, "archup-monitors.kdl")
	if err := s.writeIfChanged(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write niri monitor config: %w", err)
	}

	includeLine := "include \"archup-monitors.kdl\""
	mainPath := filepath.Join(configDir, "config.kdl")
	contents, err := s.readFileIfExists(mainPath)
	if err != nil {
		return err
	}
	if !containsLine(contents, includeLine) {
		contents = appendLine(contents, includeLine)
		if err := s.fs.WriteFile(mainPath, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("write niri config: %w", err)
		}
	}

	return nil
}

func (s *WizardService) ensureHyprlockConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	themeDir := filepath.Join(home, ".config", "themes", "bleu-theme")
	configDir := filepath.Join(home, ".config", "hypr")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create hypr config dir: %w", err)
	}

	source := filepath.Join(themeDir, "hyprlock", "hyprlock.conf")
	target := filepath.Join(configDir, "hyprlock.conf")
	data, err := s.fs.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read hyprlock config: %w", err)
	}
	if err := s.writeIfChanged(target, data, 0o644); err != nil {
		return fmt.Errorf("write hyprlock config: %w", err)
	}

	return nil
}

func (s *WizardService) ensureHypridleConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "hypr")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create hypr config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "hypridle.conf")
	content := strings.Join([]string{
		"general {",
		"    lock_cmd = hyprlock",
		"    before_sleep_cmd = loginctl lock-session",
		"    after_sleep_cmd = notify-send \"Welcome back\"",
		"}",
		"",
		"listener {",
		"    timeout = 600",
		"    on-timeout = hyprlock",
		"    on-resume = notify-send \"Unlocked\"",
		"}",
		"",
	}, "\n")

	if err := s.writeIfChanged(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write hypridle config: %w", err)
	}

	return nil
}

func (s *WizardService) ensureSDDMTheme() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	source := filepath.Join(home, ".config", "themes", "bleu-theme", "sddm", "bleu-terminal")
	if exists, err := s.fs.Exists(source); err != nil {
		return fmt.Errorf("check sddm theme source: %w", err)
	} else if !exists {
		return fmt.Errorf("sddm theme source not found: %s", source)
	}

	target := "/usr/share/sddm/themes/bleu-terminal"
	if _, err := s.cmdExec.Execute(context.Background(), "cp", "-r", source, target); err != nil {
		return fmt.Errorf("copy sddm theme: %w", err)
	}

	return nil
}

func (s *WizardService) ensureSDDMConfig(config domain.DesktopConfig) error {
	configDir := "/etc/sddm.conf.d"
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create sddm config dir: %w", err)
	}

	lines := []string{
		"[Theme]",
		"Current=bleu-terminal",
	}

	if config.AutoLogin {
		userName := resolveLoginUser()
		if userName != "" {
			lines = append(lines, "")
			lines = append(lines, "[Autologin]")
			lines = append(lines, fmt.Sprintf("User=%s", userName))
			lines = append(lines, fmt.Sprintf("Session=%s", sddmSessionName(config.Compositor)))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	configPath := filepath.Join(configDir, "archup.conf")
	if err := s.writeIfChanged(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write sddm config: %w", err)
	}

	return nil
}

func (s *WizardService) enableSDDMService() error {
	if _, err := s.cmdExec.Execute(context.Background(), "systemctl", "enable", "--now", "sddm"); err != nil {
		return fmt.Errorf("enable sddm service: %w", err)
	}
	return nil
}

func resolveLoginUser() string {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return sudoUser
	}
	current, err := osuser.Current()
	if err != nil {
		return ""
	}
	return current.Username
}

func sddmSessionName(compositor domain.Compositor) string {
	if compositor == domain.CompositorHyprland {
		return "hyprland"
	}
	return "niri"
}

func (s *WizardService) ensureWaybarConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "waybar")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create waybar config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config")
	templatePath := filepath.Join("install", "configs", "waybar", "config.json")
	configData, err := s.fs.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read waybar config template: %w", err)
	}
	if err := s.writeIfChanged(configPath, configData, 0o644); err != nil {
		return fmt.Errorf("write waybar config: %w", err)
	}

	stylePath := filepath.Join(configDir, "style.css")
	source := filepath.Join(home, ".config", "themes", "bleu-theme", "waybar", "bleu.css")
	data, err := s.fs.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read waybar style: %w", err)
	}
	if err := s.writeIfChanged(stylePath, data, 0o644); err != nil {
		return fmt.Errorf("write waybar style: %w", err)
	}

	scriptDir := filepath.Join(configDir, "scripts")
	if err := s.fs.MkdirAll(scriptDir, 0o755); err != nil {
		return fmt.Errorf("create waybar scripts dir: %w", err)
	}
	scriptPath := filepath.Join(scriptDir, "theme-toggle.sh")
	scriptSource := filepath.Join("install", "configs", "waybar", "theme-toggle.sh")
	scriptData, err := s.fs.ReadFile(scriptSource)
	if err != nil {
		return fmt.Errorf("read waybar theme toggle script: %w", err)
	}
	if err := s.writeIfChanged(scriptPath, scriptData, 0o755); err != nil {
		return fmt.Errorf("write waybar theme toggle script: %w", err)
	}

	return nil
}

func (s *WizardService) writeIfChanged(path string, data []byte, perm os.FileMode) error {
	exists, err := s.fs.Exists(path)
	if err != nil {
		return err
	}
	if exists {
		existing, err := s.fs.ReadFile(path)
		if err != nil {
			return err
		}
		if string(existing) == string(data) {
			return nil
		}
	}
	return s.fs.WriteFile(path, data, perm)
}

func (s *WizardService) applyGtkTheme() error {
	if _, err := s.cmdExec.Execute(context.Background(), "gsettings", "set", "org.gnome.desktop.interface", "color-scheme", "prefer-dark"); err != nil {
		return fmt.Errorf("set GTK color scheme: %w", err)
	}
	if _, err := s.cmdExec.Execute(context.Background(), "gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", "adw-gtk3-dark"); err != nil {
		return fmt.Errorf("set GTK theme: %w", err)
	}
	return nil
}

func (s *WizardService) enableUserServices() error {
	if err := s.ensureWaybarService(); err != nil {
		return err
	}

	if _, err := s.cmdExec.Execute(context.Background(), "systemctl", "--user", "daemon-reload"); err != nil {
		return fmt.Errorf("reload user systemd: %w", err)
	}

	if _, err := s.cmdExec.Execute(context.Background(), "systemctl", "--user", "enable", "--now", "waybar.service"); err != nil {
		return fmt.Errorf("enable waybar service: %w", err)
	}
	if _, err := s.cmdExec.Execute(context.Background(), "systemctl", "--user", "enable", "--now", "hypridle.service"); err != nil {
		return fmt.Errorf("enable hypridle service: %w", err)
	}

	return nil
}

func (s *WizardService) ensureWaybarService() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	unitDir := filepath.Join(home, ".config", "systemd", "user")
	if err := s.fs.MkdirAll(unitDir, 0o755); err != nil {
		return fmt.Errorf("create user systemd dir: %w", err)
	}

	unitPath := filepath.Join(unitDir, "waybar.service")
	unit := strings.Join([]string{
		"[Unit]",
		"Description=Waybar",
		"PartOf=graphical-session.target",
		"After=graphical-session.target",
		"",
		"[Service]",
		"ExecStart=/usr/bin/waybar",
		"Restart=on-failure",
		"",
		"[Install]",
		"WantedBy=graphical-session.target",
		"",
	}, "\n")

	if err := s.fs.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("write waybar service: %w", err)
	}

	return nil
}

func (s *WizardService) applyHyprlandConfig(home string, config domain.DesktopConfig) error {
	configDir := filepath.Join(home, ".config", "hypr")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create hypr config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "hyprland.conf")
	includeLine := "source = ~/.config/hypr/archup.conf"
	contents, err := s.readFileIfExists(configPath)
	if err != nil {
		return err
	}
	if !containsLine(contents, includeLine) {
		contents = appendLine(contents, includeLine)
	}
	if err := s.fs.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write hyprland config: %w", err)
	}

	archupPath := filepath.Join(configDir, "archup.conf")
	managed := hyprlandManagedConfig(config)
	if err := s.fs.WriteFile(archupPath, []byte(managed), 0o644); err != nil {
		return fmt.Errorf("write hyprland managed config: %w", err)
	}

	return nil
}

func (s *WizardService) applyNiriConfig(home string, config domain.DesktopConfig) error {
	configDir := filepath.Join(home, ".config", "niri")
	if err := s.fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("create niri config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.kdl")
	includeLine := "include \"archup.kdl\""
	contents, err := s.readFileIfExists(configPath)
	if err != nil {
		return err
	}
	if !containsLine(contents, includeLine) {
		contents = appendLine(contents, includeLine)
	}
	if err := s.fs.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write niri config: %w", err)
	}

	archupPath := filepath.Join(configDir, "archup.kdl")
	managed := niriManagedConfig(config)
	if err := s.fs.WriteFile(archupPath, []byte(managed), 0o644); err != nil {
		return fmt.Errorf("write niri managed config: %w", err)
	}

	return nil
}

func (s *WizardService) readFileIfExists(path string) (string, error) {
	exists, err := s.fs.Exists(path)
	if err != nil {
		return "", fmt.Errorf("check config exists: %w", err)
	}
	if !exists {
		return "", nil
	}
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}
	return string(data), nil
}

func containsLine(contents string, line string) bool {
	for _, existing := range strings.Split(contents, "\n") {
		if strings.TrimSpace(existing) == strings.TrimSpace(line) {
			return true
		}
	}
	return false
}

func appendLine(contents string, line string) string {
	trimmed := strings.TrimRight(contents, "\n")
	if trimmed == "" {
		return line + "\n"
	}
	return trimmed + "\n" + line + "\n"
}

func hyprlandManagedConfig(config domain.DesktopConfig) string {
	lines := []string{
		"# ArchUp managed config",
		"exec-once = swww-daemon",
		"exec-once = swww img ~/.config/themes/bleu-theme/wallpapers/wp-bleu-1.png",
		"exec-once = waybar",
		"exec-once = mako",
		"exec-once = hypridle",
		"exec-once = /usr/lib/polkit-kde-authentication-agent-1",
		"",
		"bind = SUPER, T, exec, foot",
		"bind = SUPER, D, exec, fuzzel",
		"bind = SUPER, Q, killactive",
		"bind = SUPER, L, exec, hyprlock",
		"",
	}

	if config.InstallCliphist {
		lines = append(lines, "exec-once = wl-paste --type text --watch cliphist store")
		lines = append(lines, "exec-once = wl-paste --type image --watch cliphist store")
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func niriManagedConfig(config domain.DesktopConfig) string {
	lines := []string{
		"// ArchUp managed config",
		"environment {",
		"    DISPLAY \":1\"",
		"    ELECTRON_OZONE_PLATFORM_HINT \"auto\"",
		"    QT_QPA_PLATFORM \"wayland\"",
		"    QT_WAYLAND_DISABLE_WINDOWDECORATION \"1\"",
		"    XDG_SESSION_TYPE \"wayland\"",
		"    XDG_CURRENT_DESKTOP \"niri\"",
		"}",
		"spawn-at-startup \"swww-daemon\"",
		"spawn-at-startup \"swww\" \"img\" \"~/.config/themes/bleu-theme/wallpapers/wp-bleu-1.png\"",
		"spawn-at-startup \"waybar\"",
		"spawn-at-startup \"mako\"",
		"spawn-at-startup \"hypridle\"",
		"spawn-at-startup \"/usr/lib/polkit-kde-authentication-agent-1\"",
		"",
		"binds {",
		"    Mod+T { spawn \"foot\"; }",
		"    Mod+D { spawn \"fuzzel\"; }",
		"    Mod+Q { close-window; }",
		"    Mod+L { spawn \"hyprlock\"; }",
		"}",
		"",
	}

	if config.InstallCliphist {
		lines = append(lines, "spawn-at-startup \"wl-paste\" \"--type\" \"text\" \"--watch\" \"cliphist\" \"store\"")
		lines = append(lines, "spawn-at-startup \"wl-paste\" \"--type\" \"image\" \"--watch\" \"cliphist\" \"store\"")
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func uniquePackages(packages []string) []string {
	seen := make(map[string]struct{}, len(packages))
	result := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}
		if _, ok := seen[pkg]; ok {
			continue
		}
		seen[pkg] = struct{}{}
		result = append(result, pkg)
	}
	return result
}

// MonitorOutput represents a detected monitor output.
type MonitorOutput struct {
	Name        string        `json:"name"`
	Enabled     bool          `json:"enabled"`
	CurrentMode *MonitorMode  `json:"current_mode"`
	Modes       []MonitorMode `json:"modes"`
}

// MonitorMode represents a display mode.
type MonitorMode struct {
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Refresh float64 `json:"refresh"`
}

// MonitorConfig represents a desired monitor configuration.
type MonitorConfig struct {
	Name    string
	Enabled bool
	Width   int
	Height  int
	Refresh float64
	PosX    int
	PosY    int
	Scale   float64
}

// DetectMonitors returns monitor outputs from wlr-randr.
func (s *WizardService) DetectMonitors(ctx context.Context) ([]MonitorOutput, error) {
	output, err := s.cmdExec.Execute(ctx, "wlr-randr", "--json")
	if err != nil {
		return nil, fmt.Errorf("detect monitors: %w", err)
	}

	var monitors []MonitorOutput
	if err := json.Unmarshal(output, &monitors); err != nil {
		return nil, fmt.Errorf("parse monitor JSON: %w", err)
	}
	return monitors, nil
}

// ApplyMonitorConfig applies monitor settings using wlr-randr.
func (s *WizardService) ApplyMonitorConfig(ctx context.Context, configs []MonitorConfig) error {
	for _, cfg := range configs {
		args := []string{"--output", cfg.Name}
		if cfg.Enabled {
			args = append(args, "--on")
			if cfg.Width > 0 && cfg.Height > 0 {
				mode := fmt.Sprintf("%dx%d", cfg.Width, cfg.Height)
				if cfg.Refresh > 0 {
					mode = fmt.Sprintf("%s@%.2fHz", mode, cfg.Refresh)
				}
				args = append(args, "--mode", mode)
			}
			args = append(args, "--pos", fmt.Sprintf("%d,%d", cfg.PosX, cfg.PosY))
			if cfg.Scale > 0 {
				args = append(args, "--scale", fmt.Sprintf("%.2f", cfg.Scale))
			}
		} else {
			args = append(args, "--off")
		}

		if _, err := s.cmdExec.Execute(ctx, "wlr-randr", args...); err != nil {
			return fmt.Errorf("apply monitor %s: %w", cfg.Name, err)
		}
	}
	return nil
}
