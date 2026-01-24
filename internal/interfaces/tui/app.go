package tui

import (
	"context"

	"github.com/bnema/archup/internal/application/services"
	"github.com/bnema/archup/internal/domain/ports"
	"github.com/bnema/archup/internal/interfaces/tui/handlers"
	"github.com/bnema/archup/internal/interfaces/tui/models"
	"github.com/bnema/archup/internal/interfaces/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

// App is the main TUI application coordinator
// It manages the installation workflow by coordinating models, views, and handlers
// with the application layer services
type App struct {
	// Core services
	installService  *services.InstallationService
	progressTracker *services.ProgressTracker

	// Infrastructure ports
	logger ports.Logger

	// UI models (using concrete types for compatibility with views/handlers)
	formModel         *models.FormModelImpl
	installationModel *models.InstallationModelImpl
	progressModel     *models.ProgressModelImpl

	// Application state
	currentScreen Screen
	ctx           context.Context
	cancel        context.CancelFunc
}

// Screen represents the current TUI screen
type Screen string

const (
	ScreenForm       Screen = "form"
	ScreenInstalling Screen = "installing"
	ScreenProgress   Screen = "progress"
	ScreenSummary    Screen = "summary"
	ScreenError      Screen = "error"
)

// NewApp creates a new TUI application
func NewApp(
	installService *services.InstallationService,
	progressTracker *services.ProgressTracker,
	logger ports.Logger,
) *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		installService:    installService,
		progressTracker:   progressTracker,
		logger:            logger,
		formModel:         models.NewFormModel(),
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
	return nil
}

// Update handles messages and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)
	case ProgressUpdateMsg:
		return a.handleProgressUpdate(msg)
	case InstallationErrorMsg:
		return a.handleInstallationError(msg)
	case InstallationCompleteMsg:
		return a.handleInstallationComplete(msg)
	}

	return a, nil
}

// View renders the current screen using the views package
func (a *App) View() string {
	switch a.currentScreen {
	case ScreenForm:
		return views.RenderForm(a.formModel)
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
	case ScreenProgress:
		// Limited input during installation
		if msg.String() == "ctrl+c" {
			return a.handleCancel()
		}
	case ScreenSummary:
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
		// Validate form and start installation
		if err := a.validateForm(); err != nil {
			a.logger.Error("Form validation failed", "error", err)
			a.currentScreen = ScreenError
			a.installationModel.SetError(err.Error())
			return a, nil
		}
		// Start installation
		return a.startInstallation()
	default:
		// Delegate to form handler
		updatedForm, cmd := handlers.HandleFormUpdate(a.formModel, msg)
		a.formModel = updatedForm
		return a, cmd
	}
}

// handleProgressUpdate processes progress events
func (a *App) handleProgressUpdate(msg ProgressUpdateMsg) (tea.Model, tea.Cmd) {
	updatedProgress, cmd := handlers.HandleProgressUpdate(a, msg, a.progressModel)
	a.progressModel = updatedProgress
	a.currentScreen = ScreenProgress
	return a, cmd
}

// handleInstallationError processes installation errors
func (a *App) handleInstallationError(msg InstallationErrorMsg) (tea.Model, tea.Cmd) {
	updatedInstall, cmd := handlers.HandleInstallationError(a, msg, a.installationModel)
	a.installationModel = updatedInstall
	a.currentScreen = ScreenError
	return a, cmd
}

// handleInstallationComplete processes installation completion
func (a *App) handleInstallationComplete(msg InstallationCompleteMsg) (tea.Model, tea.Cmd) {
	updatedInstall, cmd := handlers.HandleInstallationComplete(a, msg, a.installationModel)
	a.installationModel = updatedInstall
	a.currentScreen = ScreenSummary
	return a, cmd
}

// handleCancel cancels the installation
func (a *App) handleCancel() (tea.Model, tea.Cmd) {
	a.logger.Info("Installation cancelled by user")
	a.cancel()
	a.currentScreen = ScreenError
	a.installationModel.SetError("Installation cancelled by user")
	return a, nil
}

// validateForm validates the form data
func (a *App) validateForm() error {
	formData := a.formModel.GetData()
	if formData.Hostname == "" {
		return ErrInvalidHostname
	}
	if formData.Username == "" {
		return ErrInvalidUsername
	}
	if formData.TargetDisk == "" {
		return ErrInvalidDisk
	}
	return nil
}

// startInstallation starts the installation process
func (a *App) startInstallation() (tea.Model, tea.Cmd) {
	formData := a.formModel.GetData()
	a.currentScreen = ScreenInstalling

	// Create installation command using handler
	cmd := handlers.CreateInstallationCommand(a, formData)

	return a, tea.Batch(cmd)
}

// Close cleans up resources
func (a *App) Close() error {
	a.cancel()
	return a.installService.Close()
}
