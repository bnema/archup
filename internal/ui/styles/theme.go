package styles

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Bleu Theme - Based on the bleu-theme color palette
// https://github.com/bnema/bleu-theme
// Uses CompleteColor with TrueColor > ANSI256 > ANSI16 fallback
// Arch ISO supports 256 colors (TERM=xterm-256color)

var (
	// Background Colors
	ColorDeepNavy = lipgloss.CompleteColor{
		TrueColor: "#050a14",
		ANSI256:   "16",  // Pure black
		ANSI:      "0",   // Black
	}

	ColorDarkBlue = lipgloss.CompleteColor{
		TrueColor: "#070c16",
		ANSI256:   "233", // Very dark grey
		ANSI:      "0",   // Black
	}

	ColorDarkerBlue = lipgloss.CompleteColor{
		TrueColor: "#0a1018",
		ANSI256:   "234", // Very dark grey
		ANSI:      "8",   // Bright black (dark grey)
	}

	ColorActiveTabBlue = lipgloss.CompleteColor{
		TrueColor: "#0f1520",
		ANSI256:   "235", // Dark grey
		ANSI:      "8",   // Bright black
	}

	ColorOceanBlue = lipgloss.CompleteColor{
		TrueColor: "#2d4a6b",
		ANSI256:   "24",  // Deep blue
		ANSI:      "4",   // Blue
	}

	// Text Colors
	ColorPrimaryText = lipgloss.CompleteColor{
		TrueColor: "#e8f4f8",
		ANSI256:   "255", // Bright white
		ANSI:      "7",   // White
	}

	ColorPureWhite = lipgloss.CompleteColor{
		TrueColor: "#fefefe",
		ANSI256:   "231", // Pure white
		ANSI:      "15",  // Bright white
	}

	ColorDimmedText = lipgloss.CompleteColor{
		TrueColor: "#708090",
		ANSI256:   "102", // Medium grey
		ANSI:      "8",   // Bright black (grey)
	}

	// Accent Colors
	ColorBrightCyan = lipgloss.CompleteColor{
		TrueColor: "#00d4ff",
		ANSI256:   "45",  // Bright cyan
		ANSI:      "14",  // Bright cyan
	}

	ColorPureBlue = lipgloss.CompleteColor{
		TrueColor: "#5588cc",
		ANSI256:   "69",  // Steel blue
		ANSI:      "12",  // Bright blue
	}

	ColorLightSkyBlue = lipgloss.CompleteColor{
		TrueColor: "#87ceeb",
		ANSI256:   "117", // Light sky blue
		ANSI:      "14",  // Bright cyan
	}

	ColorSkyBlue = lipgloss.CompleteColor{
		TrueColor: "#4a7ba7",
		ANSI256:   "67",  // Steel blue
		ANSI:      "4",   // Blue
	}

	// Status Colors
	ColorSuccessGreen = lipgloss.CompleteColor{
		TrueColor: "#99FFE4",
		ANSI256:   "122", // Aquamarine green
		ANSI:      "10",  // Bright green
	}

	ColorSoftRed = lipgloss.CompleteColor{
		TrueColor: "#ff6b8a",
		ANSI256:   "204", // Light red
		ANSI:      "9",   // Bright red
	}

	ColorWarmOrange = lipgloss.CompleteColor{
		TrueColor: "#ffb347",
		ANSI256:   "215", // Light orange
		ANSI:      "11",  // Bright yellow
	}
)

// Styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorPrimaryText)

	// Title/Header styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorBrightCyan).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorPureBlue)

	// Interactive elements
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorPureWhite).
			Background(ColorOceanBlue).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorDimmedText)

	FocusedStyle = lipgloss.NewStyle().
			Foreground(ColorBrightCyan).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPureBlue)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(ColorDimmedText).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkBlue)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccessGreen).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorSoftRed).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarmOrange).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorLightSkyBlue)

	// Logo styles (for ASCII art display)
	LogoCyanStyle = lipgloss.NewStyle().
			Foreground(ColorBrightCyan)

	LogoWhiteStyle = lipgloss.NewStyle().
			Foreground(ColorPureWhite)

	// Progress/Spinner styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorBrightCyan)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorPureBlue)

	// Prompt styles
	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorBrightCyan).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorPureWhite).
			Background(ColorActiveTabBlue).
			Padding(0, 1)

	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(ColorDimmedText)

	// Box/Panel styles
	PanelStyle = lipgloss.NewStyle().
			Background(ColorDarkerBlue).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkBlue).
			Padding(1, 2)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkBlue)

	// Log/Output styles
	LogStyle = lipgloss.NewStyle().
			Foreground(ColorDimmedText).
			Background(ColorDarkerBlue).
			Padding(1, 2)

	// Help/Footer styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDimmedText).
			MarginTop(1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorPureBlue).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(ColorDimmedText)
)

// Helper functions for common styling patterns

// RenderTitle renders a section title
func RenderTitle(text string) string {
	return TitleStyle.Render(text)
}

// RenderSuccess renders success message
func RenderSuccess(text string) string {
	return SuccessStyle.Render("[OK] " + text)
}

// RenderError renders error message
func RenderError(text string) string {
	return ErrorStyle.Render("[ERROR] " + text)
}

// RenderWarning renders warning message
func RenderWarning(text string) string {
	return WarningStyle.Render("[WARN] " + text)
}

// RenderInfo renders info message
func RenderInfo(text string) string {
	return InfoStyle.Render("[INFO] " + text)
}

// RenderKeyValue renders a key-value pair
func RenderKeyValue(key, value string) string {
	return KeyStyle.Render(key+": ") + lipgloss.NewStyle().Foreground(ColorPrimaryText).Render(value)
}

// RenderPanel renders content in a panel
func RenderPanel(content string) string {
	return PanelStyle.Render(content)
}

// RenderHelp renders help text with key bindings
func RenderHelp(bindings map[string]string) string {
	var help string
	for key, desc := range bindings {
		help += KeyStyle.Render(key) + " " + DescStyle.Render(desc) + "  "
	}
	return HelpStyle.Render(help)
}

// HuhTheme returns a custom huh.Theme matching the bleu color scheme
// Uses ThemeBase16 as foundation for maximum TTY compatibility (QEMU console, etc)
func HuhTheme() *huh.Theme {
	// Start with Base16 theme for limited terminal support
	theme := huh.ThemeBase16()

	// Customize with bleu theme colors using ANSI 16 codes
	// ANSI Color mapping for bleu theme:
	// 0=black, 4=blue, 6=cyan, 7=white, 8=grey, 10=bright green, 12=bright blue, 14=bright cyan, 15=bright white

	// Form and Group styles
	theme.Group.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)       // Bright cyan
	theme.Group.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // Bright blue

	// Focused field styles
	theme.Focused.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)     // Bright cyan
	theme.Focused.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))          // Bright blue
	theme.Focused.SelectSelector = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).SetString("> ")
	theme.Focused.Option = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))                // White
	theme.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).SetString("> ")
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))       // Bright cyan
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).SetString("✓ ") // Bright green
	theme.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))      // White
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString("○ ") // Grey

	// Button styles with explicit background colors for TTY compatibility
	// Active/focused button (currently selected - e.g., "Yes" when on Yes)
	theme.Focused.FocusedButton = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).  // Bright white text
		Background(lipgloss.Color("4")).   // Blue background
		Padding(0, 2).
		Bold(true)

	// Inactive/blurred button (not selected - e.g., "No" when on Yes)
	theme.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).   // White text
		Background(lipgloss.Color("0")).   // Black background
		Padding(0, 2)

	// Blurred field styles (when entire field loses focus)
	theme.Blurred.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))                 // Grey
	theme.Blurred.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))           // Grey
	theme.Blurred.SelectSelector = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).SetString("> ")
	theme.Blurred.Option = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))                // Grey

	// Override buttons in Blurred state too (inherited from Focused in Base16)
	theme.Blurred.FocusedButton = theme.Focused.FocusedButton
	theme.Blurred.BlurredButton = theme.Focused.BlurredButton

	// Text input styles
	theme.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))     // Bright cyan
	theme.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Grey
	theme.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true) // Bright cyan

	// Error styles
	theme.Focused.ErrorIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).SetString("✗ ") // Bright red
	theme.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)         // Bright red

	return theme
}
