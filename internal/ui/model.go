package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/phases"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/bnema/archup/internal/ui/styles"
)

// ModelState represents the current UI state
type ModelState int

const (
	StateShowLogo ModelState = iota
	StatePreflightForm
	StateDiskForm
	StateOptionsForm
	StateConfirmation
	StateExecuting
	StateComplete
	StateError
)

// Model is the main Bubbletea application model
type Model struct {
	state        ModelState
	orchestrator *phases.Orchestrator
	config       *config.Config
	currentForm  *huh.Form
	output       *components.OutputViewer
	version      string
	width        int
	height       int
	err          error
}

// NewModel creates a new installer model
func NewModel(orchestrator *phases.Orchestrator, cfg *config.Config, version string) *Model {
	return &Model{
		state:        StateShowLogo,
		orchestrator: orchestrator,
		config:       cfg,
		version:      version,
		width:        80,
		height:       24,
		output:       components.NewOutputViewer(80, 20, 1000),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.output.SetSize(msg.Width, msg.Height-10)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			switch m.state {
			case StateShowLogo:
				// Move to preflight form
				m.state = StatePreflightForm
				m.currentForm = CreatePreflightForm(m.config)
				return m, m.currentForm.Init()
			}
		}

	case phaseCompleteMsg:
		switch {
		case msg.err != nil:
			m.err = msg.err
			m.state = StateError
			return m, tea.Quit
		}

		// Check if all phases complete
		switch {
		case m.orchestrator.IsComplete():
			m.state = StateComplete
			return m, tea.Quit
		default:
			// Continue with next phase
			return m, m.executeNextPhase()
		}

	case phases.ProgressUpdate:
		// Handle progress updates from phases
		switch {
		case msg.OutputLine != "":
			m.output.AddLine(msg.OutputLine)
		case msg.Step != "":
			m.output.AddLine(fmt.Sprintf("[%s] %s", msg.PhaseName, msg.Step))
		case msg.IsError:
			m.output.AddLine(fmt.Sprintf("[ERROR] %s", msg.ErrorMsg))
		case msg.IsComplete:
			m.output.AddLine(fmt.Sprintf("[COMPLETE] %s", msg.Step))
		}
		return m, waitForProgress(m.orchestrator.ProgressChannel())
	}

	// Update current form if showing one
	switch m.state {
	case StatePreflightForm, StateDiskForm, StateOptionsForm:
		switch {
		case m.currentForm != nil:
			form, cmd := m.currentForm.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.currentForm = f

				// Check if form is completed
				switch {
				case m.currentForm.State == huh.StateCompleted:
					return m, m.handleFormComplete()
				}
			}
			return m, cmd
		}
	case StateExecuting:
		// Update output viewer
		return m, m.output.Update(msg)
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	switch m.state {
	case StateShowLogo:
		logo := NewLogo(m.version)
		return logo.RenderCentered(m.width) + "\n\n" +
			centerText(WelcomeTagline, m.width) + "\n\n" +
			centerText(styles.HelpStyle.Render("Press ENTER to start installation"), m.width)

	case StatePreflightForm, StateDiskForm, StateOptionsForm:
		switch {
		case m.currentForm != nil:
			return m.currentForm.View()
		}
		return "Loading form..."

	case StateConfirmation:
		return m.renderConfirmation()

	case StateExecuting:
		return styles.TitleStyle.Render("Installing ArchUp...") + "\n\n" + m.output.View()

	case StateError:
		return styles.RenderError(fmt.Sprintf("Error: %v", m.err))

	case StateComplete:
		return styles.RenderSuccess("Installation complete! You can now reboot.")

	default:
		return "Unknown state"
	}
}

// handleFormComplete handles form completion and moves to next state
func (m *Model) handleFormComplete() tea.Cmd {
	switch m.state {
	case StatePreflightForm:
		m.state = StateDiskForm
		m.currentForm = CreateDiskSelectionForm(m.config)
		return m.currentForm.Init()

	case StateDiskForm:
		m.state = StateOptionsForm
		m.currentForm = CreateOptionsForm(m.config)
		return m.currentForm.Init()

	case StateOptionsForm:
		// Check if encryption password needed
		switch m.config.EncryptionType {
		case config.EncryptionLUKS, config.EncryptionLUKSLVM:
			m.currentForm = CreateEncryptionPasswordForm(m.config)
			return m.currentForm.Init()
		}

		// Move to confirmation
		m.state = StateConfirmation
		return nil

	default:
		// Encryption password form completed
		// If password is empty, use user password
		switch {
		case m.config.EncryptPassword == "":
			m.config.EncryptPassword = m.config.UserPassword
		}

		// Move to confirmation
		m.state = StateConfirmation
		return nil
	}
}

// renderConfirmation renders the confirmation screen
func (m *Model) renderConfirmation() string {
	s := styles.TitleStyle.Render("Confirm Installation") + "\n\n"
	s += fmt.Sprintf("Hostname: %s\n", m.config.Hostname)
	s += fmt.Sprintf("Username: %s\n", m.config.Username)
	s += fmt.Sprintf("Disk: %s\n", m.config.TargetDisk)
	s += fmt.Sprintf("Kernel: %s\n", m.config.KernelChoice)
	s += fmt.Sprintf("Encryption: %s\n", m.config.EncryptionType)
	s += "\n"
	s += styles.WarningStyle.Render("WARNING: This will erase all data on "+m.config.TargetDisk) + "\n\n"
	s += styles.HelpStyle.Render("Press ENTER to continue, Ctrl+C to cancel")
	return s
}

// executeNextPhase executes the next pending phase
func (m *Model) executeNextPhase() tea.Cmd {
	return func() tea.Msg {
		err := m.orchestrator.ExecuteNext()
		return phaseCompleteMsg{err: err}
	}
}

// waitForProgress waits for progress updates
func waitForProgress(progressChan <-chan phases.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		return <-progressChan
	}
}

// phaseCompleteMsg signals phase execution completion
type phaseCompleteMsg struct {
	err error
}

// centerText centers text within a given width
func centerText(text string, width int) string {
	textLen := lipgloss.Width(text)
	switch {
	case textLen >= width:
		return text
	}
	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}
