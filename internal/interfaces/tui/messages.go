package tui

import "github.com/bnema/archup/internal/application/dto"

// InstallationErrorMsg is sent when an error occurs during installation
type InstallationErrorMsg struct {
	Err error
}

// InstallationCompleteMsg is sent when installation completes successfully
type InstallationCompleteMsg struct {
	Duration int // seconds
}

// FormSubmitMsg is sent when the form is submitted
type FormSubmitMsg struct {
	Data FormData
}

// ScreenChangeMsg is sent to change the current screen
type ScreenChangeMsg struct {
	Screen Screen
}

// ProgressUpdateMsg is sent when installation progress updates
type ProgressUpdateMsg struct {
	Update *dto.ProgressUpdate
}
