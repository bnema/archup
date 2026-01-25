package domain

// DesktopConfig represents wizard selections for desktop setup.
type DesktopConfig struct {
	Compositor      Compositor
	EnableSDDM      bool
	AutoLogin       bool
	AutoLoginUser   string
	InstallCliphist bool
}

// NewDesktopConfig returns defaults for a new wizard run.
func NewDesktopConfig() DesktopConfig {
	return DesktopConfig{
		Compositor:      CompositorNiri,
		EnableSDDM:      true,
		AutoLogin:       false,
		AutoLoginUser:   "",
		InstallCliphist: false,
	}
}
