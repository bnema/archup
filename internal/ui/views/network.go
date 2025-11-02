package views

import (
	"fmt"

	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderNetworkCheck renders the network connectivity check screen
func RenderNetworkCheck(m model.UI) string {
	spinner := m.Spinner()
	spinnerView := spinner.View()

	var content string

	// If network check is still running
	if !m.NetworkCheckDone() {
		content = CenterText(spinnerView+" Checking network connectivity...", m.Width()) + "\n\n" +
			CenterText(styles.HelpStyle.Render("Please wait..."), m.Width())
	} else if m.NetworkErr() != nil {
		// Network check failed
		errorMsg := fmt.Sprintf("Network error: %v", m.NetworkErr())
		content = CenterText(styles.ErrorStyle.Render("✗ Network Check Failed"), m.Width()) + "\n\n" +
			CenterText(errorMsg, m.Width()) + "\n\n" +
			CenterText(styles.HelpStyle.Render("Press ENTER to retry or CTRL+C to exit"), m.Width())
	} else {
		// Network check passed
		content = CenterText(styles.SuccessStyle.Render("✓ Network Check Passed"), m.Width()) + "\n\n" +
			CenterText(styles.HelpStyle.Render("Press ENTER to continue"), m.Width())
	}

	return styles.CenteredContainerStyle.Render(content)
}
