package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderConfirm renders confirmation screen.
func RenderConfirm(config domain.DesktopConfig) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	cliphistStatus := "no"
	if config.InstallCliphist {
		cliphistStatus = "yes"
	}

	sddmStatus := "no"
	if config.EnableSDDM {
		sddmStatus = "yes"
	}

	autoLoginStatus := "no"
	if config.AutoLogin {
		autoLoginStatus = "yes"
	}

	b.WriteString("\n")
	b.WriteString(title.Render("Confirm Selections"))
	b.WriteString("\n\n")
	b.WriteString("Compositor: " + string(config.Compositor) + "\n")
	b.WriteString("SDDM: " + sddmStatus + "\n")
	b.WriteString("Auto-login: " + autoLoginStatus + "\n")
	b.WriteString("Cliphist: " + cliphistStatus + "\n\n")
	b.WriteString(muted.Render("Enter Install • B Back"))

	return b.String()
}
