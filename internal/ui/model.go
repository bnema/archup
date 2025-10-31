package ui

import (
	"fmt"
	"time"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/phases"
	"github.com/bnema/archup/internal/system"
	"github.com/bnema/archup/internal/ui/components"
	"github.com/bnema/archup/internal/ui/styles"
	"github.com/bnema/archup/internal/ui/views"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// ModelState represents the current UI state
type ModelState int

const (
	StateShowLogo ModelState = iota
	StateNetworkCheck
	StatePreflightForm
	StateDiskForm
	StateOptionsForm
	StateAMDPStatePrompt
	StateConfirmation
	StateExecuting
	StateComplete
	StateResultsSuccess
	StateResultsError
	StateError
)

// Model is the main Bubbletea application model
type Model struct {
	state            ModelState
	orchestrator     *phases.Orchestrator
	config           *config.Config
	currentForm      *huh.Form
	formBuilder      *components.FormBuilder
	output           *components.OutputViewer
	logger           *logger.Logger
	system           *system.System
	spinner          spinner.Model
	version          string
	width            int
	height           int
	err              error
	cpuInfo          *system.CPUInfo
	installStartTime time.Time
	installEndTime   time.Duration
	networkCheckDone bool
	networkErr       error
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
		system:       system.NewSystem(log.Slog()),
		spinner:      s,
		version:      version,
		width:        80,
		height:       24,
		output:       components.NewOutputViewer(80, 20, 1000),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	m.installStartTime = time.Now()
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
				m.logger.Info("Moving to network check")
				m.state = StateNetworkCheck
				return m, m.checkNetwork()

			case StateNetworkCheck:
				// Handle retry if network check failed
				if m.networkErr != nil {
					m.logger.Info("Retrying network check")
					return m, m.checkNetwork()
				}
				// Network check passed, move to preflight form
				m.logger.Info("Moving to preflight form")
				m.state = StatePreflightForm
				m.currentForm = CreatePreflightForm(m.config, m.formBuilder, m.system)
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
				m.currentForm = CreatePreflightForm(m.config, m.formBuilder, m.system)
				return m, m.currentForm.Init()

			case StateOptionsForm:
				m.state = StateDiskForm
				m.currentForm = CreateDiskSelectionForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()

			case StateAMDPStatePrompt:
				m.state = StateOptionsForm
				m.currentForm = CreateOptionsForm(m.config, m.formBuilder)
				return m, m.currentForm.Init()

			case StateConfirmation:
				// Go back to AMD P-State prompt if we have AMD CPU, otherwise options
				if m.cpuInfo != nil && m.cpuInfo.Vendor == system.CPUVendorAMD && len(m.cpuInfo.AMDPStateModes) > 0 {
					m.state = StateAMDPStatePrompt
					m.currentForm = CreateAMDPStateForm(m.config, m.cpuInfo, m.formBuilder)
				} else {
					m.state = StateOptionsForm
					m.currentForm = CreateOptionsForm(m.config, m.formBuilder)
				}
				return m, m.currentForm.Init()
			}
		}

	case networkCheckMsg:
		m.networkCheckDone = true
		m.networkErr = msg.err
		if msg.err != nil {
			m.logger.Warn("Network check failed", "error", msg.err)
			return m, nil
		}
		// Network check passed - automatically continue to preflight form
		m.logger.Info("Network check passed - continuing automatically")
		m.state = StatePreflightForm
		m.currentForm = CreatePreflightForm(m.config, m.formBuilder, m.system)
		return m, m.currentForm.Init()

	case phaseCompleteMsg:
		if msg.err != nil {
			m.logger.Error("Phase execution failed", "error", msg.err)
			m.err = msg.err
			m.installEndTime = time.Since(m.installStartTime)
			m.state = StateResultsError
			return m, tea.Quit
		}

		// Check if all phases complete
		switch {
		case m.orchestrator.IsComplete():
			m.logger.Info("All phases completed successfully")
			m.installEndTime = time.Since(m.installStartTime)
			m.state = StateResultsSuccess
			return m, tea.Quit
		default:
			m.logger.Info("Continuing to next phase")
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
	case StatePreflightForm, StateDiskForm, StateOptionsForm, StateAMDPStatePrompt:
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
				if m.currentForm.State == huh.StateCompleted {
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
		return views.RenderWelcome(m)

	case StateNetworkCheck:
		return views.RenderNetworkCheck(m)

	case StatePreflightForm, StateDiskForm, StateOptionsForm, StateAMDPStatePrompt:
		return views.RenderForm(m)

	case StateConfirmation:
		return views.RenderConfirmation(m)

	case StateExecuting:
		return views.RenderExecuting(m)

	case StateResultsSuccess:
		phaseDurations := m.orchestrator.GetPhaseDurations()
		return views.RenderSuccess(m, m.installEndTime, phaseDurations)

	case StateResultsError:
		phaseName := "unknown"
		if phase := m.orchestrator.CurrentPhase(); phase != nil {
			phaseName = phase.Name()
		}
		return views.RenderError(m, m.err, phaseName, m.installEndTime)

	case StateComplete, StateError:
		if m.err != nil {
			return styles.RenderError(fmt.Sprintf("Error: %v", m.err))
		}
		return styles.RenderSuccess("Installation complete!")

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
		if m.config.UseSamePasswordForEncryption {
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

		// Detect CPU and check if we need to show AMD P-State prompt
		cpuInfo, err := system.DetectCPUInfo()
		if err == nil {
			m.cpuInfo = cpuInfo
			// Show AMD P-State prompt only for AMD CPUs with P-State support
			if cpuInfo.Vendor == system.CPUVendorAMD && len(cpuInfo.AMDPStateModes) > 0 {
				m.logger.Info("AMD CPU detected - showing P-State configuration",
					"zen_gen", cpuInfo.AMDZenGen.Label,
					"modes", len(cpuInfo.AMDPStateModes))
				m.state = StateAMDPStatePrompt
				m.currentForm = CreateAMDPStateForm(m.config, cpuInfo, m.formBuilder)
				return m.currentForm.Init()
			}
		}

		// Move directly to confirmation if not AMD or detection failed
		m.state = StateConfirmation
		return nil

	case StateAMDPStatePrompt:
		m.logger.Info("AMD P-State configuration completed", "mode", m.config.AMDPState)
		m.state = StateConfirmation
		return nil

	default:
		// Move to confirmation
		m.state = StateConfirmation
		return nil
	}
}

// Getter methods for views package
func (m *Model) Width() int { return m.width }
func (m *Model) Version() string { return m.version }
func (m *Model) Config() *config.Config { return m.config }
func (m *Model) CurrentForm() *huh.Form { return m.currentForm }
func (m *Model) Spinner() spinner.Model { return m.spinner }
func (m *Model) Output() *components.OutputViewer { return m.output }
func (m *Model) RenderPhaseHeader() string { return m.renderPhaseHeader() }
func (m *Model) NetworkCheckDone() bool { return m.networkCheckDone }
func (m *Model) NetworkErr() error { return m.networkErr }

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

// networkCheckMsg signals network check completion
type networkCheckMsg struct {
	err error
}

// checkNetwork runs network connectivity check in background
func (m *Model) checkNetwork() tea.Cmd {
	return func() tea.Msg {
		err := system.CheckNetworkConnectivity()
		return networkCheckMsg{err: err}
	}
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
	case StateAMDPStatePrompt:
		stepNum = "4"
		stepName = "AMD P-State Configuration"
	default:
		return ""
	}

	// Total steps depends on whether AMD prompt is shown
	totalSteps := "3"
	if m.state == StateAMDPStatePrompt {
		totalSteps = "4"
	}

	header := fmt.Sprintf("ArchUp Install - Step %s/%s: %s", stepNum, totalSteps, stepName)
	return styles.TitleStyle.Render(header)
}
