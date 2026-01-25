package views

import (
	"strings"

	"github.com/bnema/archup/internal/wizard/domain"
	"github.com/charmbracelet/lipgloss"
)

// RenderSDDM renders the SDDM selection screen.
func RenderSDDM(config domain.DesktopConfig) string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	sddmStatus := "disabled"
	if config.EnableSDDM {
		sddmStatus = "enabled"
	}
	autoLoginStatus := "off"
	if config.AutoLogin {
		autoLoginStatus = "on"
	}

	b.WriteString("\n")
	b.WriteString(title.Render("Display Manager"))
	b.WriteString("\n\n")
	b.WriteString("SDDM: " + sddmStatus + "\n")
	b.WriteString("Auto-login: " + autoLoginStatus + "\n\n")
	b.WriteString(muted.Render("Space Toggle SDDM • A Toggle Auto-login • Enter Next"))

	return b.String()
}
