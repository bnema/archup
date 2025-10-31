package views

import (
	"fmt"

	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderConfirmation renders the confirmation screen (left-aligned)
func RenderConfirmation(m model.UI) string {
	cfg := m.Config()

	s := styles.TitleStyle.Render("Confirm Installation") + "\n\n"
	s += fmt.Sprintf("Hostname: %s\n", cfg.Hostname)
	s += fmt.Sprintf("Username: %s\n", cfg.Username)
	s += fmt.Sprintf("Disk: %s\n", cfg.TargetDisk)
	s += fmt.Sprintf("Kernel: %s\n", cfg.KernelChoice)
	s += fmt.Sprintf("Encryption: %s\n", cfg.EncryptionType)

	if cfg.AMDPState != "" {
		s += fmt.Sprintf("AMD P-State: %s\n", cfg.AMDPState)
	}

	s += "\n"
	s += styles.WarningStyle.Render("WARNING: This will erase all data on "+cfg.TargetDisk) + "\n\n"
	s += styles.HelpStyle.Render("Press ENTER to continue, Ctrl+C to cancel")

	return styles.FormContainerStyle.Render(s)
}
