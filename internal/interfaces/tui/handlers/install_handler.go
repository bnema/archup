package handlers

import (
	"github.com/bnema/archup/internal/application/dto"
	"github.com/bnema/archup/internal/interfaces/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

// NOTE: This package uses AppContext interface to avoid importing parent tui package
// This breaks the circular dependency: tui → tui/handlers → tui
// Messages are typed by the caller (app.go), not here

// HandleProgressUpdate processes progress update messages
// msg should be ProgressUpdateMsg with an Update field (*dto.ProgressUpdate)
func HandleProgressUpdate(app AppContext, msg interface{}, progressModel *models.ProgressModelImpl) (*models.ProgressModelImpl, tea.Cmd) {
	// Cast to a type with Update field
	// This avoids importing the tui package which would create a cycle
	type updateMsg struct {
		Update interface{}
	}

	if m, ok := msg.(updateMsg); ok {
		// Cast Update to *dto.ProgressUpdate
		if update, ok := m.Update.(*dto.ProgressUpdate); ok {
			progressModel.UpdateProgress(update)
		}
	}
	return progressModel, nil
}

// HandleInstallationError processes installation error messages
// Signature matches what app.go calls with InstallationErrorMsg types
func HandleInstallationError(app AppContext, msg interface{}, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	// Cast msg to InstallationErrorMsg type (from app.go caller context)
	if errMsg, ok := msg.(interface{ GetErr() error }); ok {
		err := errMsg.GetErr()
		app.GetLogger().Error("Installation error", "error", err)
		installModel.SetError(err.Error())
	}
	return installModel, nil
}

// HandleInstallationComplete processes installation completion messages
// Signature matches what app.go calls with InstallationCompleteMsg types
func HandleInstallationComplete(app AppContext, msg interface{}, installModel *models.InstallationModelImpl) (*models.InstallationModelImpl, tea.Cmd) {
	app.GetLogger().Info("Installation completed successfully")
	installModel.SetComplete()
	return installModel, nil
}

// CreateInstallationCommand creates an installation command from form data
// This delegates to the service layer instead of handling directly
// Accepts AppContext interface to avoid circular imports
func CreateInstallationCommand(app AppContext, formData models.FormData) tea.Cmd {
	return func() tea.Msg {
		// Start installation via service
		if err := app.GetInstallService().Start(
			app.GetContext(),
			formData.Hostname,
			formData.Username,
			formData.TargetDisk,
			formData.EncryptionType,
		); err != nil {
			app.GetLogger().Error("Failed to start installation", "error", err)
			// Return error directly - app.go will handle it
			return err
		}

		// Subscribe to progress updates
		progressChan := app.GetProgressTracker().Subscribe()
		go func() {
			for {
				select {
				case <-progressChan:
					// Progress updates are broadcast to all subscribers
					// App will receive them through its own subscription
				case <-app.GetContext().Done():
					app.GetLogger().Info("Installation context cancelled")
					return
				}
			}
		}()

		return nil
	}
}
