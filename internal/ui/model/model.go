package model

import (
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/huh"
)

// UI is an interface that views use to access model data
type UI interface {
	Width() int
	Version() string
	Config() *config.Config
	CurrentForm() *huh.Form
	Spinner() spinner.Model
	Output() *components.OutputViewer
	RenderPhaseHeader() string
	NetworkCheckDone() bool
	NetworkErr() error
}
