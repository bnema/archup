package views

import (
	"github.com/bnema/archup/internal/ui/model"
	"github.com/bnema/archup/internal/ui/styles"
)

// RenderWelcome renders the welcome logo screen (centered)
func RenderWelcome(m model.UI) string {
	logo := NewLogo(m.Version())
	content := logo.RenderCentered(m.Width()) + "\n\n\n" +
		CenterText(WelcomeTagline, m.Width()) + "\n\n" +
		CenterText(styles.HelpStyle.Render("Press ENTER to start installation"), m.Width())

	return styles.CenteredContainerStyle.Render(content)
}
