package views

import (
	"strings"

	"github.com/bnema/archup/internal/interfaces/tui/models"
	legacysystem "github.com/bnema/archup/internal/system"
	"github.com/charmbracelet/lipgloss"
)

// RenderAMDPStateSelection renders the AMD P-State selection screen.
func RenderAMDPStateSelection(am *models.AMDPStateModelImpl) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	b.WriteString("\n")
	b.WriteString(title.Render("AMD P-State Configuration"))
	b.WriteString("\n\n")

	if am == nil || am.CPUInfo() == nil {
		b.WriteString(info.Render("Detecting CPU..."))
		b.WriteString("\n\n")
		b.WriteString(info.Render("This screen will appear only for AMD CPUs."))
		return b.String()
	}

	cpuInfo := am.CPUInfo()
	if cpuInfo.Vendor != legacysystem.CPUVendorAMD {
		b.WriteString(info.Render("Non-AMD CPU detected. Skipping AMD P-State configuration."))
		return b.String()
	}

	if cpuInfo.AMDZenGen != nil {
		b.WriteString(info.Render("Detected: "))
		b.WriteString(cpuInfo.AMDZenGen.Label)
		b.WriteString("\n")
	}
	if cpuInfo.ModelName != "" {
		b.WriteString(info.Render("CPU: "))
		b.WriteString(cpuInfo.ModelName)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(info.Render("Note: Requires CPPC enabled in UEFI (AMD CBS > NBIO > SMU > CPPC)"))
	b.WriteString("\n\n")

	options := am.Options()
	if len(options) == 0 {
		b.WriteString(info.Render("No AMD P-State modes available for this CPU."))
		return b.String()
	}

	for i, option := range options {
		prefix := "  "
		style := lipgloss.NewStyle()
		label := option.Label
		if option.Recommended {
			label = label + " (recommended)"
		}
		if i == am.SelectedIndex() {
			prefix = "> "
			style = active
		}
		b.WriteString(style.Render(prefix + label))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(info.Render("↑/↓ navigate • enter confirm • esc back • ctrl+c quit"))

	return b.String()
}
