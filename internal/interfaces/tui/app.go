package tui

import (
	"context"
	"os"

	apphandlers "github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/application/services"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/domain/system"
	"github.com/bnema/archup/internal/interfaces/tui/handlers"
	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/bnema/archup/internal/interfaces/tui/views"
	legacysystem "github.com/bnema/archup/internal/system"
	tea "github.com/charmbracelet/bubbletea"
)

// App is the main TUI application coordinator
// It manages the installation workflow by coordinating models, views, and handlers
// with the application layer services
type App struct {
	// Core services
	installService  *services.InstallationService
	progressTracker *services.ProgressTracker
	gpuHandler      *apphandlers.GPUHandler

	// Infrastructure ports
	logger ports.Logger

	// UI models (using concrete types for compatibility with views/handlers)
	formModel         *models.FormModelImpl
	diskModel         *models.DiskModelImpl
	encryptionModel   *models.EncryptionModelImpl
	kernelModel       *models.KernelModelImpl
	amdPstateModel    *models.AMDPStateModelImpl
	gpuModel          *models.GPUModelImpl
	reposModel        *models.ReposModelImpl
	dankLinuxModel    *models.DankLinuxModelImpl
	installationModel *models.InstallationModelImpl
	progressModel     *models.ProgressModelImpl

	// Application state
	currentScreen Screen
	ctx           context.Context
	cancel        context.CancelFunc

	// Validated form data - stored after form validation passes
	formData models.FormData

	// Program reference for sending messages from goroutines
	program *tea.Program
}

// Screen represents the current TUI screen
type Screen string

const (
	ScreenForm       Screen = "form"
	ScreenDisk       Screen = "disk"
	ScreenEncryption Screen = "encryption"
	ScreenKernel     Screen = "kernel"
	ScreenAMDPState  Screen = "amd-pstate"
	ScreenGPU        Screen = "gpu"
	ScreenRepos      Screen = "repos"
	ScreenDankLinux  Screen = "danklinux"
	ScreenInstalling Screen = "installing"
	ScreenProgress   Screen = "progress"
	ScreenSummary    Screen = "summary"
	ScreenError      Screen = "error"
)

// NewApp creates a new TUI application
func NewApp(
	installService *services.InstallationService,
	progressTracker *services.ProgressTracker,
	gpuHandler *apphandlers.GPUHandler,
	logger ports.Logger,
) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		installService:    installService,
		progressTracker:   progressTracker,
		gpuHandler:        gpuHandler,
		logger:            logger,
		formModel:         models.NewFormModel(),
		diskModel:         models.NewDiskModel(),
		encryptionModel:   models.NewEncryptionModel(),
		kernelModel:       models.NewKernelModel(),
		amdPstateModel:    models.NewAMDPStateModel(),
		gpuModel:          models.NewGPUModel(),
		reposModel:        models.NewReposModel(),
		dankLinuxModel:    models.NewDankLinuxModel(),
		installationModel: models.NewInstallationModel(),
		progressModel:     models.NewProgressModel(),
		currentScreen:     ScreenForm,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Getter methods for handlers to access app resources

// GetLogger returns the logger
func (a *App) GetLogger() ports.Logger {
	return a.logger
}

// GetInstallService returns the installation service
func (a *App) GetInstallService() *services.InstallationService {
	return a.installService
}

// GetProgressTracker returns the progress tracker
func (a *App) GetProgressTracker() *services.ProgressTracker {
	return a.progressTracker
}

// GetContext returns the app context
func (a *App) GetContext() context.Context {
	return a.ctx
}

// Init is called when the program starts
func (a *App) Init() tea.Cmd {
	a.logger.Debug("TUI app initialized")
	return tea.Batch(
		a.formModel.Init(),
		a.detectTimezoneCmd(),
	)
}

// Update handles messages and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)
	case handlers.ProgressUpdateMsg:
		return a.handleProgressUpdate(msg)
	case handlers.InstallationErrorMsg:
		return a.handleInstallationError(msg)
	case handlers.InstallationCompleteMsg:
		return a.handleInstallationComplete(msg)
	case CPUDetectedMsg:
		return a.handleCPUDetected(msg)
	case GPUDetectedMsg:
		return a.handleGPUDetected(msg)
	case TimezoneDetectedMsg:
		return a.handleTimezoneDetected(msg)
	case DisksDetectedMsg:
		return a.handleDisksDetected(msg)
	}

	return a, nil
}

// View renders the current screen using the views package
func (a *App) View() string {
	switch a.currentScreen {
	case ScreenForm:
		return views.RenderForm(a.formModel)
	case ScreenDisk:
		return views.RenderDiskSelection(a.diskModel)
	case ScreenEncryption:
		return views.RenderEncryptionSelection(a.encryptionModel)
	case ScreenKernel:
		return views.RenderKernelSelection(a.kernelModel)
	case ScreenAMDPState:
		return views.RenderAMDPStateSelection(a.amdPstateModel)
	case ScreenGPU:
		return views.RenderGPUSelection(a.gpuModel)
	case ScreenRepos:
		return views.RenderReposSelection(a.reposModel)
	case ScreenDankLinux:
		return views.RenderDankLinuxSelection(a.dankLinuxModel)
	case ScreenInstalling, ScreenProgress:
		return views.RenderProgress(a.progressModel)
	case ScreenSummary:
		return views.RenderSummary(a.installationModel)
	case ScreenError:
		return views.RenderError(a.installationModel.GetError())
	default:
		return views.RenderForm(a.formModel)
	}
}

// handleKeyMsg handles keyboard input
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.currentScreen {
	case ScreenForm:
		return a.handleFormInput(msg)
	case ScreenDisk:
		return a.handleDiskInput(msg)
	case ScreenEncryption:
		return a.handleEncryptionInput(msg)
	case ScreenKernel:
		return a.handleKernelInput(msg)
	case ScreenAMDPState:
		return a.handleAMDPStateInput(msg)
	case ScreenGPU:
		return a.handleGPUInput(msg)
	case ScreenRepos:
		return a.handleReposInput(msg)
	case ScreenDankLinux:
		return a.handleDankLinuxInput(msg)
	case ScreenProgress:
		// Limited input during installation
		if msg.String() == "ctrl+c" {
			return a.handleCancel()
		}
	case ScreenSummary:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "r":
			a.unmountAndReboot()
			return a, nil
		}
	case ScreenError:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	return a, nil
}

// handleFormInput processes form screen input
func (a *App) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "enter":
		a.formModel.ExtractData()
		// Validate form and go to disk selection
		if err := a.validateForm(); err != nil {
			a.logger.Warn("Form validation failed", "error", err)
			// Show error on form, don't leave screen
			a.formModel.SetError(err)
			return a, nil
		}
		// Clear any previous error and store validated data in App
		a.formModel.SetError(nil)
		a.formData = a.formModel.GetData()
		a.logger.Debug("Form data stored", "password_len", len(a.formData.UserPassword))
		return a.startDiskSelection()
	default:
		// Delegate to form handler
		updatedForm, cmd := handlers.HandleFormUpdate(a.formModel, msg)
		a.formModel = updatedForm
		return a, cmd
	}
}

// handleProgressUpdate processes progress events
func (a *App) handleProgressUpdate(msg handlers.ProgressUpdateMsg) (tea.Model, tea.Cmd) {
	updatedProgress, cmd := handlers.HandleProgressUpdate(a, msg, a.progressModel)
	a.progressModel = updatedProgress
	// Don't override summary or error screens with late progress messages
	if a.currentScreen != ScreenSummary && a.currentScreen != ScreenError {
		a.currentScreen = ScreenProgress
	}
	return a, cmd
}

// handleInstallationError processes installation errors
func (a *App) handleInstallationError(msg handlers.InstallationErrorMsg) (tea.Model, tea.Cmd) {
	updatedInstall, cmd := handlers.HandleInstallationError(a, msg, a.installationModel)
	a.installationModel = updatedInstall
	a.currentScreen = ScreenError
	return a, cmd
}

// handleInstallationComplete processes installation completion
func (a *App) handleInstallationComplete(msg handlers.InstallationCompleteMsg) (tea.Model, tea.Cmd) {
	updatedInstall, cmd := handlers.HandleInstallationComplete(a, msg, a.installationModel)
	a.installationModel = updatedInstall
	a.currentScreen = ScreenSummary
	return a, cmd
}

// handleGPUDetected updates GPU model with detection result.
func (a *App) handleGPUDetected(msg GPUDetectedMsg) (tea.Model, tea.Cmd) {
	a.gpuModel.SetDetectedGPU(msg.GPU)
	return a, nil
}

// handleCancel cancels the installation
func (a *App) handleCancel() (tea.Model, tea.Cmd) {
	a.logger.Info("Installation cancelled by user")
	a.cancel()
	a.currentScreen = ScreenError
	a.installationModel.SetError("Installation cancelled by user")
	return a, nil
}

func (a *App) unmountAndReboot() {
	mountPoint := "/mnt"
	check := legacysystem.RunSimple("mountpoint", "-q", mountPoint)
	if check.ExitCode == 0 {
		if err := legacysystem.Unmount(a.logger.LogPath(), mountPoint); err != nil {
			a.logger.Warn("Failed to unmount before reboot", "error", err)
			a.installationModel.SetNotice("Failed to unmount /mnt. Please run: umount -R /mnt")
			return
		}
	}

	if _, err := os.Stat("/dev/mapper/cryptroot"); err == nil {
		result := legacysystem.RunLogged(a.logger.LogPath(), "cryptsetup", "close", "cryptroot")
		if result.Error != nil {
			a.logger.Warn("Failed to close encrypted volume", "error", result.Error)
		}
	}

	a.logger.Info("Rebooting system")
	_ = legacysystem.RunLogged(a.logger.LogPath(), "reboot")
}

func (a *App) handleGPUInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "up", "shift+tab":
		a.gpuModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.gpuModel.MoveDown()
		return a, nil
	case "enter":
		selected := a.gpuModel.SelectedOption()
		// Update stored form data directly
		a.formData.GPUVendor = string(selected.Vendor)
		a.formData.GPUDrivers = append([]string{}, selected.Drivers...)
		return a.startReposSelection()
	}

	return a, nil
}

func (a *App) handleDankLinuxInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		return a.startReposSelection()
	case "up", "shift+tab":
		a.dankLinuxModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.dankLinuxModel.MoveDown()
		return a, nil
	case "enter":
		a.formData.InstallDankLinux = a.dankLinuxModel.SelectedOption().Value
		return a.startInstallation()
	case "q", "ctrl+c":
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) startReposSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenRepos
	return a, nil
}

func (a *App) handleReposInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc", "backspace":
		return a.startGPUSelection()
	case "up":
		a.reposModel.MoveUp()
		return a, nil
	case "down":
		a.reposModel.MoveDown()
		return a, nil
	case "tab":
		a.reposModel.NextSection()
		return a, nil
	case "shift+tab":
		a.reposModel.PrevSection()
		return a, nil
	case "enter":
		a.formData.AURHelper = a.reposModel.SelectedAURHelper()
		a.formData.EnableChaotic = a.reposModel.SelectedChaoticEnabled()
		a.currentScreen = ScreenDankLinux
		return a, nil
	}
	return a, nil
}

func (a *App) handleKernelInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		return a.startDiskSelection()
	case "up", "shift+tab":
		a.kernelModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.kernelModel.MoveDown()
		return a, nil
	case "enter":
		selected := a.kernelModel.SelectedOption()
		a.formData.KernelVariant = selected.Package
		return a.startAMDPStateSelection()
	}

	return a, nil
}

func (a *App) handleCPUDetected(msg CPUDetectedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		a.logger.Warn("CPU detection reported error", "error", msg.Err)
	}

	a.amdPstateModel.SetCPUInfo(msg.CPU)

	// Always include microcode when CPU vendor is known
	if msg.CPU != nil && msg.CPU.Vendor != "" && msg.CPU.Vendor != legacysystem.CPUVendorUnknown {
		a.formData.Microcode = true
	}

	if !a.amdPstateModel.ShouldPrompt() {
		a.formData.AMDPState = ""
		a.formData.KernelParamsExtra = ""
		return a.startGPUSelection()
	}

	return a, nil
}

func (a *App) handleAMDPStateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		return a.startKernelSelection()
	case "up", "shift+tab":
		a.amdPstateModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.amdPstateModel.MoveDown()
		return a, nil
	case "enter":
		selected := a.amdPstateModel.SelectedOption()
		mode := string(selected.Mode)
		a.formData.AMDPState = mode
		a.formData.KernelParamsExtra = ""

		cpuInfo := a.amdPstateModel.CPUInfo()
		switch {
		case mode == "":
			// No selection
		case cpuInfo != nil && cpuInfo.AMDZenGen != nil:
			a.formData.KernelParamsExtra = legacysystem.GetAMDPStateKernelParams(cpuInfo.AMDZenGen, selected.Mode)
		default:
			a.formData.KernelParamsExtra = "amd_pstate=" + mode
		}

		return a.startGPUSelection()
	}

	return a, nil
}

// validateForm validates the form data using domain validation
func (a *App) validateForm() error {
	formData := a.formModel.GetData()

	// Use domain validation for all fields
	if err := system.ValidateHostname(formData.Hostname); err != nil {
		return err
	}

	if err := system.ValidateUsername(formData.Username); err != nil {
		return err
	}

	if err := system.ValidatePassword(formData.UserPassword); err != nil {
		return err
	}

	if err := system.ValidateTimezone(formData.Timezone); err != nil {
		return err
	}

	if err := system.ValidateLocale(formData.Locale); err != nil {
		return err
	}

	if err := system.ValidateKeymap(formData.Keymap); err != nil {
		return err
	}

	// TargetDisk is selected in disk selection screen, not validated here
	return nil
}

// startInstallation starts the installation process
func (a *App) startInstallation() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenInstalling
	a.logger.Debug("Starting installation", "password_len", len(a.formData.UserPassword))

	// Create installation command using stored form data
	cmd := handlers.CreateInstallationCommand(a, a.formData)

	return a, tea.Batch(cmd)
}

func (a *App) startAMDPStateSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenAMDPState
	return a, a.detectCPUCmd()
}

func (a *App) startKernelSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenKernel
	a.kernelModel.SetSelectedPackage(a.formData.KernelVariant)
	return a, nil
}

func (a *App) detectCPUCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := legacysystem.DetectCPUInfo()
		if err != nil {
			a.logger.Warn("CPU detection failed", "error", err)
		}
		return CPUDetectedMsg{CPU: info, Err: err}
	}
}

func (a *App) startGPUSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenGPU
	return a, a.detectGPUCmd()
}

func (a *App) detectGPUCmd() tea.Cmd {
	return func() tea.Msg {
		gpu, err := a.gpuHandler.Detect(a.ctx)
		if err != nil {
			a.logger.Warn("GPU detection failed", "error", err)
		}
		return GPUDetectedMsg{GPU: gpu}
	}
}

func (a *App) startDiskSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenDisk
	return a, a.detectDisksCmd()
}

func (a *App) detectDisksCmd() tea.Cmd {
	return func() tea.Msg {
		disks, err := legacysystem.ListDisks()
		return DisksDetectedMsg{Disks: disks, Err: err}
	}
}

func (a *App) detectTimezoneCmd() tea.Cmd {
	return func() tea.Msg {
		tz := legacysystem.DetectTimezoneSimple()
		return TimezoneDetectedMsg{Timezone: tz}
	}
}

func (a *App) handleDiskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "up", "shift+tab":
		a.diskModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.diskModel.MoveDown()
		return a, nil
	case "enter":
		selected := a.diskModel.SelectedOption()
		if selected.Path == "" {
			// No disk selected
			a.currentScreen = ScreenError
			a.installationModel.SetError("No disk selected")
			return a, nil
		}
		// Update stored form data directly
		a.formData.TargetDisk = selected.Path
		return a.startEncryptionSelection()
	}
	return a, nil
}

func (a *App) startEncryptionSelection() (tea.Model, tea.Cmd) {
	a.currentScreen = ScreenEncryption
	return a, nil
}

func (a *App) handleEncryptionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		return a.startDiskSelection()
	case "up", "shift+tab":
		a.encryptionModel.MoveUp()
		return a, nil
	case "down", "tab":
		a.encryptionModel.MoveDown()
		return a, nil
	case "enter":
		a.formData.EncryptionType = a.encryptionModel.SelectedOption().Value
		return a.startKernelSelection()
	}
	return a, nil
}

func (a *App) handleTimezoneDetected(msg TimezoneDetectedMsg) (tea.Model, tea.Cmd) {
	if msg.Timezone != "" {
		a.logger.Info("Timezone detected", "timezone", msg.Timezone)
		a.formModel.SetTimezone(msg.Timezone)
	}
	return a, nil
}

func (a *App) handleDisksDetected(msg DisksDetectedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		a.logger.Warn("Disk detection failed", "error", msg.Err)
		a.diskModel.SetError(msg.Err)
	} else {
		a.diskModel.SetDisks(msg.Disks)
	}
	return a, nil
}

// Close cleans up resources
func (a *App) Close() error {
	a.cancel()
	return a.installService.Close()
}

// SetProgram sets the tea.Program reference for sending messages from goroutines
func (a *App) SetProgram(p *tea.Program) {
	a.program = p
}

// GetProgram returns the tea.Program reference
func (a *App) GetProgram() *tea.Program {
	return a.program
}
