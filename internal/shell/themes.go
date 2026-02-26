package shell

// ThemeResult holds the result of theme application
type ThemeResult struct {
	Applied  int
	Failed   []string
	Warnings []string
}

// applyThemes is a no-op. Theme setup is handled by cli-tools.sh at first boot.
func (c *Configurator) applyThemes(username, userHome string) ThemeResult {
	return ThemeResult{
		Applied:  0,
		Failed:   []string{},
		Warnings: []string{},
	}
}
