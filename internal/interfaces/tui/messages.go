package tui

import (
	"github.com/bnema/archup/internal/domain/system"
	legacysystem "github.com/bnema/archup/internal/system"
)

// Note: ProgressUpdateMsg, InstallationErrorMsg, InstallationCompleteMsg
// are defined in handlers package to avoid circular imports

// FormSubmitMsg is sent when the form is submitted
type FormSubmitMsg struct {
	Data FormData
}

// ScreenChangeMsg is sent to change the current screen
type ScreenChangeMsg struct {
	Screen Screen
}

// GPUDetectedMsg is sent when GPU detection completes
type GPUDetectedMsg struct {
	GPU *system.GPU
}

// TimezoneDetectedMsg is sent when timezone detection completes
type TimezoneDetectedMsg struct {
	Timezone string
}

// DisksDetectedMsg is sent when disk detection completes
type DisksDetectedMsg struct {
	Disks []legacysystem.Disk
	Err   error
}
