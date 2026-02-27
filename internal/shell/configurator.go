package shell

// ConfigResult holds the result of shell configuration
type ConfigResult struct {
	ThemesApplied int
	ThemesFailed  []string
	Warnings      []string
}

// Configurator handles shell and bash configuration setup
type Configurator struct{}

// NewConfigurator creates a new shell configurator
func NewConfigurator() *Configurator {
	return &Configurator{}
}

// Configure is a no-op. Shell configuration is handled by cli-tools.sh at first boot.
func (c *Configurator) Configure(username, userHome string) (ConfigResult, error) {
	return ConfigResult{
		ThemesApplied: 0,
		ThemesFailed:  []string{},
		Warnings:      []string{"Shell configuration deferred to cli-tools.sh on first boot"},
	}, nil
}
