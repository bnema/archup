package handlers

import (
	"context"

	"github.com/bnema/archup/internal/application/services"
	"github.com/bnema/archup/internal/domain/ports"
	tea "github.com/charmbracelet/bubbletea"
)

// AppContext defines the interface for accessing app resources from handlers
// This avoids circular dependencies while allowing handlers to access needed services
type AppContext interface {
	// GetLogger returns the logger
	GetLogger() ports.Logger

	// GetInstallService returns the installation service
	GetInstallService() *services.InstallationService

	// GetProgressTracker returns the progress tracker
	GetProgressTracker() *services.ProgressTracker

	// GetContext returns the app context
	GetContext() context.Context

	// GetProgram returns the tea.Program for sending messages from goroutines
	GetProgram() *tea.Program
}
