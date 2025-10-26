package ui

import (
	"fmt"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/phases"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/bnema/archup/internal/ui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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
	formBuilder  *components.FormBuilder
	output       *components.OutputViewer
	logger       *logger.Logger
	spinner      spinner.Model
	version      string
	width        int
	height       int
	err          error
}

// NewModel creates a new installer model
func NewModel(orchestrator *phases.Orchestrator, cfg *config.Config, log *logger.Logger, version string) *Model {
	log.Info("UI Model created", "version", version)

	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = styles.SpinnerStyle

	return &Model{
		state:        StateShowLogo,
		orchestrator: orchestrator,
		config:       cfg,
		formBuilder:  components.NewFormBuilder(false, 80), // Initial width, will update on WindowSizeMsg
		logger:       log,
		spinner:      s,
		version:      version,
		width:        80,
		height:       24,
		output:       components.NewOutputViewer(80, 20, 1000),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.output.SetSize(msg.Width, msg.Height-10)
		// Update formBuilder width for responsive forms
		m.formBuilder = components.NewFormBuilder(false, msg.Width)

	case tea.KeyMsg:
		m.logger.Info("Key pressed", "key", msg.String(), "state", m.state)

		switch msg.String() {
		case "ctrl+c":
			m.logger.Info("User cancelled installation")
			return m, tea.Quit

		case "enter":
			switch m.state {
			case StateShowLogo:
				m.logger.Info("Moving to preflight form")
				m.state = StatePreflightForm
				m.currentForm = CreatePreflightForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()

			case StateConfirmation:
				m.logger.Info("User confirmed installation - starting execution")
				m.state = StateExecuting
				return m, m.executeNextPhase()
			}

		case "backspace", "esc":
			m.logger.Info("User navigating back", "from_state", m.state)
			switch m.state {
			case StateDiskForm:
				m.state = StatePreflightForm
				m.currentForm = CreatePreflightForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()

			case StateOptionsForm:
				m.state = StateDiskForm
				m.currentForm = CreateDiskSelectionForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()

			case StateConfirmation:
				m.state = StateOptionsForm
				m.currentForm = CreateOptionsForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()
			}
		}

	case phaseCompleteMsg:
		switch {
		case msg.err != nil:
			m.logger.Error("Phase execution failed", "error", msg.err)
			m.err = msg.err
			m.state = StateError
			return m, tea.Quit
		}

		// Check if all phases complete
		switch {
		case m.orchestrator.IsComplete():
			m.logger.Info("All phases completed successfully")
			m.state = StateComplete
			return m, tea.Quit
		default:
			m.logger.Info("Continuing to next phase")
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
			// Don't pass WindowSizeMsg to form to preserve our max width
			if _, isWindowSize := msg.(tea.WindowSizeMsg); isWindowSize {
				return m, nil
			}

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
		// Update spinner and output viewer
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.output.Update(msg))
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	switch m.state {
	case StateShowLogo:
		logo := NewLogo(m.version)
		return logo.RenderCentered(m.width) + "\n\n\n" +
			centerText(WelcomeTagline, m.width) + "\n\n" +
			centerText(styles.HelpStyle.Render("Press ENTER to start installation"), m.width)

	case StatePreflightForm, StateDiskForm, StateOptionsForm:
		switch {
		case m.currentForm != nil:
			header := centerText(m.renderPhaseHeader(), m.width)
			return header + "\n\n" + m.currentForm.View()
		}
		return "Loading form..."

	case StateConfirmation:
		return m.renderConfirmation()

	case StateExecuting:
		title := m.spinner.View() + " Installation in progress..."
		return styles.TitleStyle.Render(title) + "\n\n" + m.output.View()

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
	m.logger.Info("Form completed", "state", m.state)

	switch m.state {
	case StatePreflightForm:
		m.logger.Info("Preflight form completed",
			"hostname", m.config.Hostname,
			"username", m.config.Username,
			"use_same_password_for_encryption", m.config.UseSamePasswordForEncryption)

		// Handle encryption password setting based on checkbox
		switch {
		case m.config.UseSamePasswordForEncryption:
			m.config.EncryptPassword = m.config.UserPassword
		}

		m.state = StateDiskForm
		m.currentForm = CreateDiskSelectionForm(m.config, m.formBuilder)
		return m.currentForm.Init()

	case StateDiskForm:
		m.logger.Info("Disk selection completed", "disk", m.config.TargetDisk)
		m.state = StateOptionsForm
		m.currentForm = CreateOptionsForm(m.config, m.formBuilder)
		return m.currentForm.Init()

	case StateOptionsForm:
		m.logger.Info("Options form completed",
			"kernel", m.config.KernelChoice,
			"encryption", m.config.EncryptionType)
		// Move directly to confirmation
		// Encryption password is already handled in preflight form
		m.state = StateConfirmation
		return nil

	default:
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
	// Execute the phase in the background AND start listening for progress updates
	return tea.Batch(
		func() tea.Msg {
			err := m.orchestrator.ExecuteNext()
			return phaseCompleteMsg{err: err}
		},
		waitForProgress(m.orchestrator.ProgressChannel()),
	)
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

// centerText centers text within a given width using lipgloss
func centerText(text string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(text)
}


// renderPhaseHeader renders phase progress header
func (m *Model) renderPhaseHeader() string {
	var stepNum, stepName string

	switch m.state {
	case StatePreflightForm:
		stepNum = "1"
		stepName = "System Configuration"
	case StateDiskForm:
		stepNum = "2"
		stepName = "Disk Selection"
	case StateOptionsForm:
		stepNum = "3"
		stepName = "Installation Options"
	default:
		return ""
	}

	header := fmt.Sprintf("ArchUp Install - Step %s/3: %s", stepNum, stepName)
	return styles.TitleStyle.Render(header)
}
