package tui

import (
	"context"
	"time"

	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/wizard/application/services"
	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/bnema/archup/internal/wizard/interfaces/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// App is the wizard TUI model.
type App struct {
	service *services.WizardService
	logger  ports.Logger
	config  domain.DesktopConfig
	screen  Screen
	errMsg  string

	compositorIndex int
	monitors        []services.MonitorOutput
	monitorConfigs  []services.MonitorConfig
	monitorSelected int
	monitorModeIdx  map[string]int
	monitorDirty    bool
	monitorApplySeq int
	monitorApplying bool
	monitorErr      string
}

// Screen represents wizard screens.
type Screen string

const (
	ScreenWelcome    Screen = "welcome"
	ScreenCompositor Screen = "compositor"
	ScreenSDDM       Screen = "sddm"
	ScreenOptional   Screen = "optional"
	ScreenConfirm    Screen = "confirm"
	ScreenInstall    Screen = "install"
	ScreenMonitor    Screen = "monitor"
	ScreenConfig     Screen = "config"
	ScreenComplete   Screen = "complete"
)

// NewApp creates a new wizard TUI app.
func NewApp(service *services.WizardService, logger ports.Logger) *App {
	return &App{
		service: service,
		logger:  logger,
		config:  domain.NewDesktopConfig(),
		screen:  ScreenWelcome,
	}
}

// Init initializes the TUI.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles input events.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case wizardInstallMsg:
		if msg.err != nil {
			a.errMsg = msg.err.Error()
			a.screen = ScreenComplete
			return a, nil
		}
		a.screen = ScreenMonitor
		return a, a.detectMonitorsCmd()
	case wizardMonitorMsg:
		if msg.err != nil {
			a.errMsg = msg.err.Error()
			a.screen = ScreenComplete
			return a, nil
		}
		a.monitors = msg.monitors
		a.monitorConfigs = buildMonitorConfigs(msg.monitors)
		a.monitorModeIdx = buildMonitorModeIndex(msg.monitors)
		a.monitorSelected = 0
		a.monitorDirty = false
		a.monitorApplying = false
		a.monitorErr = ""
		return a, nil
	case monitorApplyTickMsg:
		if msg.seq != a.monitorApplySeq || !a.monitorDirty {
			return a, nil
		}
		a.monitorDirty = false
		a.monitorApplying = true
		return a, a.applyMonitorConfigCmd(msg.seq)
	case monitorAppliedMsg:
		if msg.seq != a.monitorApplySeq {
			return a, nil
		}
		a.monitorApplying = false
		a.monitorErr = ""
		if msg.err != nil {
			a.monitorErr = msg.err.Error()
			a.monitorDirty = true
			return a, a.scheduleMonitorApply()
		}
		return a, nil
	case wizardConfigMsg:
		if msg.err != nil {
			a.errMsg = msg.err.Error()
		}
		a.screen = ScreenComplete
		return a, nil
	case tea.KeyMsg:
		return a.handleKey(msg)
	}
	return a, nil
}

// View renders the wizard screen.
func (a *App) View() string {
	switch a.screen {
	case ScreenWelcome:
		return views.RenderWelcome(a.config)
	case ScreenCompositor:
		return views.RenderCompositor(a.config, a.compositorIndex)
	case ScreenSDDM:
		return views.RenderSDDM(a.config)
	case ScreenOptional:
		return views.RenderOptional(a.config)
	case ScreenConfirm:
		return views.RenderConfirm(a.config)
	case ScreenInstall:
		return views.RenderInstall()
	case ScreenMonitor:
		return views.RenderMonitors(a.monitors, a.monitorConfigs, a.monitorSelected, a.monitorApplying, a.monitorDirty, a.monitorErr)
	case ScreenConfig:
		return views.RenderApplyConfig()
	case ScreenComplete:
		return views.RenderComplete(a.config, a.errMsg)
	default:
		return views.RenderWelcome(a.config)
	}
}

func (a *App) startCmd() tea.Cmd {
	return func() tea.Msg {
		err := a.service.Install(context.Background(), a.config)
		if err != nil {
			a.logger.Error("Wizard start failed", "error", err)
		}
		return wizardInstallMsg{err: err}
	}
}

func (a *App) detectMonitorsCmd() tea.Cmd {
	return func() tea.Msg {
		monitors, err := a.service.DetectMonitors(context.Background())
		if err != nil {
			a.logger.Error("Monitor detection failed", "error", err)
		}
		return wizardMonitorMsg{monitors: monitors, err: err}
	}
}

func (a *App) applyConfigCmd() tea.Cmd {
	return func() tea.Msg {
		err := a.service.ApplyConfigWithMonitors(a.config, a.monitorConfigs)
		if err != nil {
			a.logger.Error("Wizard config apply failed", "error", err)
		}
		return wizardConfigMsg{err: err}
	}
}

func (a *App) applyMonitorConfigCmd(seq int) tea.Cmd {
	configs := append([]services.MonitorConfig{}, a.monitorConfigs...)
	return func() tea.Msg {
		err := a.service.ApplyMonitorConfig(context.Background(), configs)
		return monitorAppliedMsg{seq: seq, err: err}
	}
}

func (a *App) scheduleMonitorApply() tea.Cmd {
	a.monitorApplySeq++
	seq := a.monitorApplySeq
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		return monitorApplyTickMsg{seq: seq}
	})
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	}

	switch a.screen {
	case ScreenWelcome:
		if msg.String() == "enter" {
			a.screen = ScreenCompositor
		}
	case ScreenCompositor:
		switch msg.String() {
		case "up", "shift+tab":
			a.moveCompositor(-1)
		case "down", "tab":
			a.moveCompositor(1)
		case "enter":
			a.applyCompositorSelection()
			a.screen = ScreenSDDM
		}
	case ScreenSDDM:
		switch msg.String() {
		case " ":
			a.config.EnableSDDM = !a.config.EnableSDDM
			if !a.config.EnableSDDM {
				a.config.AutoLogin = false
				return a, nil
			}
		case "a":
			if a.config.EnableSDDM {
				a.config.AutoLogin = !a.config.AutoLogin
			}
		case "enter":
			a.screen = ScreenOptional
		}
	case ScreenOptional:
		switch msg.String() {
		case " ":
			a.config.InstallCliphist = !a.config.InstallCliphist
		case "enter":
			a.screen = ScreenConfirm
		}
	case ScreenConfirm:
		switch msg.String() {
		case "enter":
			a.screen = ScreenInstall
			return a, a.startCmd()
		case "b":
			a.screen = ScreenOptional
		}
	case ScreenInstall:
		return a, nil
	case ScreenMonitor:
		return a.handleMonitorKey(msg)
	case ScreenConfig:
		return a, nil
	case ScreenComplete:
		if msg.String() == "enter" {
			return a, tea.Quit
		}
	}

	return a, nil
}

func (a *App) moveCompositor(delta int) {
	options := compositorOptions()
	count := len(options)
	if count == 0 {
		return
	}
	idx := (a.compositorIndex + delta) % count
	if idx < 0 {
		idx = count - 1
	}
	a.compositorIndex = idx
}

func (a *App) applyCompositorSelection() {
	options := compositorOptions()
	if len(options) == 0 {
		return
	}
	if a.compositorIndex < 0 || a.compositorIndex >= len(options) {
		a.compositorIndex = 0
	}
	a.config.Compositor = options[a.compositorIndex]
}

func compositorOptions() []domain.Compositor {
	return []domain.Compositor{domain.CompositorNiri, domain.CompositorHyprland}
}

func buildMonitorConfigs(monitors []services.MonitorOutput) []services.MonitorConfig {
	configs := make([]services.MonitorConfig, 0, len(monitors))
	for _, monitor := range monitors {
		cfg := services.MonitorConfig{
			Name:    monitor.Name,
			Enabled: monitor.Enabled,
			Scale:   1,
		}
		if monitor.CurrentMode != nil {
			cfg.Width = monitor.CurrentMode.Width
			cfg.Height = monitor.CurrentMode.Height
			cfg.Refresh = monitor.CurrentMode.Refresh
		}
		configs = append(configs, cfg)
	}
	return configs
}

func buildMonitorModeIndex(monitors []services.MonitorOutput) map[string]int {
	index := make(map[string]int, len(monitors))
	for _, monitor := range monitors {
		idx := 0
		if monitor.CurrentMode != nil {
			idx = findModeIndex(monitor.Modes, *monitor.CurrentMode)
		}
		index[monitor.Name] = idx
	}
	return index
}

func findModeIndex(modes []services.MonitorMode, current services.MonitorMode) int {
	for i, mode := range modes {
		if mode.Width == current.Width && mode.Height == current.Height && floatEquals(mode.Refresh, current.Refresh) {
			return i
		}
	}
	return 0
}

func floatEquals(a, b float64) bool {
	if a == b {
		return true
	}
	if a > b {
		return a-b < 0.05
	}
	return b-a < 0.05
}

func (a *App) cycleMonitorMode(delta int) {
	selected := a.selectedMonitor()
	if selected == nil {
		return
	}
	var modes []services.MonitorMode
	for _, monitor := range a.monitors {
		if monitor.Name == selected.Name {
			modes = monitor.Modes
			break
		}
	}
	if len(modes) == 0 {
		return
	}
	current := a.monitorModeIdx[selected.Name]
	current = (current + delta) % len(modes)
	if current < 0 {
		current = len(modes) - 1
	}
	a.monitorModeIdx[selected.Name] = current
	mode := modes[current]
	selected.Width = mode.Width
	selected.Height = mode.Height
	selected.Refresh = mode.Refresh
}

func (a *App) handleMonitorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.screen = ScreenConfig
		return a, a.applyConfigCmd()
	case "up", "shift+tab":
		a.monitorSelected = max(0, a.monitorSelected-1)
	case "down", "tab":
		if len(a.monitorConfigs) > 0 {
			a.monitorSelected = min(len(a.monitorConfigs)-1, a.monitorSelected+1)
		}
	case " ":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().Enabled = !a.selectedMonitor().Enabled
			return a, a.markMonitorDirty()
		}
	case "m":
		if a.selectedMonitor() != nil {
			a.cycleMonitorMode(1)
			return a, a.markMonitorDirty()
		}
	case "shift+m":
		if a.selectedMonitor() != nil {
			a.cycleMonitorMode(-1)
			return a, a.markMonitorDirty()
		}
	case "left", "h":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().PosX -= 50
			return a, a.markMonitorDirty()
		}
	case "right", "l":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().PosX += 50
			return a, a.markMonitorDirty()
		}
	case "k":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().PosY -= 50
			return a, a.markMonitorDirty()
		}
	case "j":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().PosY += 50
			return a, a.markMonitorDirty()
		}
	case "r":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().PosX = 0
			a.selectedMonitor().PosY = 0
			return a, a.markMonitorDirty()
		}
	case "+", "=":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().Scale = clampScale(a.selectedMonitor().Scale + 0.1)
			return a, a.markMonitorDirty()
		}
	case "-", "_":
		if a.selectedMonitor() != nil {
			a.selectedMonitor().Scale = clampScale(a.selectedMonitor().Scale - 0.1)
			return a, a.markMonitorDirty()
		}
	}

	return a, nil
}

func (a *App) markMonitorDirty() tea.Cmd {
	a.monitorDirty = true
	return a.scheduleMonitorApply()
}

func (a *App) selectedMonitor() *services.MonitorConfig {
	if a.monitorSelected < 0 || a.monitorSelected >= len(a.monitorConfigs) {
		return nil
	}
	return &a.monitorConfigs[a.monitorSelected]
}

func clampScale(scale float64) float64 {
	if scale < 0.5 {
		return 0.5
	}
	if scale > 3.0 {
		return 3.0
	}
	return scale
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type wizardInstallMsg struct {
	err error
}

type wizardMonitorMsg struct {
	monitors []services.MonitorOutput
	err      error
}

type wizardConfigMsg struct {
	err error
}

type monitorApplyTickMsg struct {
	seq int
}

type monitorAppliedMsg struct {
	seq int
	err error
}
