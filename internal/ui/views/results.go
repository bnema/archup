package views

import (
	"fmt"
	"time"

	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderSuccess renders the success screen with timing breakdown (centered)
func RenderSuccess(m model.UI, duration time.Duration, phaseDurations map[string]time.Duration) string {
	cfg := m.Config()

	s := styles.SuccessStyle.Render("[OK] Installation Complete!") + "\n\n"
	s += "Installation Summary:\n"
	s += fmt.Sprintf("  Hostname: %s\n", cfg.Hostname)
	s += fmt.Sprintf("  Disk: %s\n", cfg.TargetDisk)
	s += fmt.Sprintf("  Kernel: %s\n", cfg.KernelChoice)
	s += fmt.Sprintf("  Total Time: %s\n\n", formatDuration(duration))

	s += "Next Steps:\n"
	s += "  1. Remove installation media\n"
	s += "  2. Reboot the system\n"
	s += "  3. First-boot setup will run automatically\n\n"
	s += styles.HelpStyle.Render("Press ENTER to exit")

	return styles.CenteredContainerStyle.Render(s)
}

// RenderError renders the error screen with log tail (centered)
func RenderError(m model.UI, err error, phaseName string, duration time.Duration) string {
	s := styles.ErrorStyle.Render("[KO] Installation Failed") + "\n\n"
	s += fmt.Sprintf("Phase: %s\n", phaseName)
	s += fmt.Sprintf("Error: %v\n", err)
	s += fmt.Sprintf("Duration before failure: %s\n\n", formatDuration(duration))

	s += "Recent log entries:\n"
	// TODO: Add log tail here once LogViewer is implemented

	s += "\nLog file: " + m.Config().LogPath + "\n"
	s += "Report issues: https://github.com/bnema/archup/issues\n\n"
	s += styles.HelpStyle.Render("Press ENTER to exit")

	return styles.CenteredContainerStyle.Render(s)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
}
