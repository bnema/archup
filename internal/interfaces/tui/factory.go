package tui

import "github.com/bnema/archup/internal/interfaces/tui/models"

// NewFormModel creates a new form model
func NewFormModel() FormModel {
	return models.NewFormModel()
}

// NewInstallationModel creates a new installation model
func NewInstallationModel() InstallationModel {
	return models.NewInstallationModel()
}

// NewProgressModel creates a new progress model
func NewProgressModel() ProgressModel {
	return models.NewProgressModel()
}

