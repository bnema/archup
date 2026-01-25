package views

import (
	"fmt"
	"strings"

	"github.com/bnema/archup/internal/wizard/application/services"
	"github.com/charmbracelet/lipgloss"
)

// RenderMonitors renders monitor configuration info.
func RenderMonitors(monitors []services.MonitorOutput, configs []services.MonitorConfig, selected int, applying bool, dirty bool, errMsg string) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	b.WriteString("\n")
	b.WriteString(title.Render("Monitor Configuration"))
	b.WriteString("\n\n")

	if len(monitors) == 0 {
		b.WriteString("No monitors detected.\n\n")
		b.WriteString(muted.Render("Press Enter to continue."))
		return b.String()
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	for i, monitor := range monitors {
		status := "off"
		if monitor.Enabled {
			status = "on"
		}
		label := fmt.Sprintf("%s (%s)", monitor.Name, status)
		if i == selected {
			label = selectedStyle.Render("› " + label)
		} else {
			label = "  " + label
		}
		b.WriteString(label)
		modeWidth, modeHeight, modeRefresh := 0, 0, 0.0
		if i < len(configs) {
			modeWidth = configs[i].Width
			modeHeight = configs[i].Height
			modeRefresh = configs[i].Refresh
		} else if monitor.CurrentMode != nil {
			modeWidth = monitor.CurrentMode.Width
			modeHeight = monitor.CurrentMode.Height
			modeRefresh = monitor.CurrentMode.Refresh
		}
		if modeWidth > 0 && modeHeight > 0 {
			b.WriteString(fmt.Sprintf(" - %dx%d @ %.2fHz", modeWidth, modeHeight, modeRefresh))
		}
		b.WriteString("\n")

		if i < len(configs) {
			cfg := configs[i]
			b.WriteString(muted.Render(fmt.Sprintf("    pos=%d,%d scale=%.2f", cfg.PosX, cfg.PosY, cfg.Scale)))
			b.WriteString("\n")
		}
	}

	if errMsg != "" {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(errMsg))
	}

	b.WriteString("\n")
	if applying {
		b.WriteString(muted.Render("Applying changes..."))
	} else if dirty {
		b.WriteString(muted.Render("Changes pending..."))
	} else {
		b.WriteString(muted.Render("Detected via wlr-randr --json"))
	}
	b.WriteString("\n")
	b.WriteString(muted.Render("↑↓ Select • H/L X-pos • J/K Y-pos • +/- Scale • M Cycle Mode • Space Toggle • Enter Next"))

	return b.String()
}
