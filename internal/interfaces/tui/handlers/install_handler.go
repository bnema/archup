package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/bnema/archup/internal/interfaces/tui"
	"github.com/bnema/archup/internal/interfaces/tui/models"
)

// HandleProgressUpdate processes progress update messages
func HandleProgressUpdate(app *tui.App, msg tui.ProgressUpdateMsg, progressModel *models.ProgressModelImpl) (*models.ProgressModelImpl, tea.Cmd) {
	progressModel.UpdateProgress(msg.Update)
	return progressModel, nil
}

// HandleInstallationError processes installation error messages
func HandleInstallationError(app *tui.App, msg tui.InstallationErrorMsg, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	app.GetLogger().Error("Installation error", "error", msg.Err)
	installModel.SetError(msg.Err.Error())
	return installModel, nil
}

// HandleInstallationComplete processes installation completion messages
func HandleInstallationComplete(app *tui.App, msg tui.InstallationCompleteMsg, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	app.GetLogger().Info("Installation completed successfully")
	installModel.SetComplete()
	return installModel, nil
}

// CreateInstallationCommand creates an installation command from form data
// This delegates to the service layer instead of handling directly
func CreateInstallationCommand(app *tui.App, formData models.FormData) tea.Cmd {
	return func() tea.Msg {
		// Start installation via service
		if err := app.GetInstallService().Start(
			app.GetContext(),
			formData.Hostname,
			formData.Username,
			formData.TargetDisk,
			formData.EncryptionType,
		); err != nil {
			return tui.InstallationErrorMsg{Err: err}
		}

		// Subscribe to progress updates and forward them as messages
		progressChan := app.GetProgressTracker().Subscribe()
		for {
			select {
			case update := <-progressChan:
				// Forward progress update to TUI
				return tui.ProgressUpdateMsg{Update: update}
			case <-app.GetContext().Done():
				return tui.InstallationErrorMsg{Err: app.GetContext().Err()}
			}
		}
	}
}
