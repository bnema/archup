package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bnema/archup/internal/cleanup"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/phases"
	"github.com/bnema/archup/internal/ui"
)

// version is set via ldflags at build time
// Example: go build -ldflags "-X main.version=v0.3.0"
var version = "dev"

func main() {
	// Parse command-line flags
	var (
		showVersion = flag.Bool("version", false, "Show version and exit")
		doCleanup   = flag.Bool("cleanup", false, "Clean up installation artifacts and exit")
		dryRun      = flag.Bool("dry-run", false, "Show TUI but don't execute commands")
	)
	flag.Parse()

	// Handle --version flag
	if *showVersion {
		fmt.Printf("archup-installer %s\n", version)
		os.Exit(0)
	}

	// Create logger with dry-run mode
	log, err := logger.New(config.DefaultLogPath, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Handle --cleanup flag
	if *doCleanup {
		if err := cleanup.Run(log); err != nil {
			log.Error("Cleanup failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Log dry-run mode if enabled
	if *dryRun {
		log.Info("Running in DRY-RUN mode - commands will be logged but not executed")
	}

	// Load or create config
	cfg := config.NewConfig()

	// Initialize orchestrator
	orchestrator := phases.NewOrchestrator(cfg, config.DefaultLogPath)

	// Register all phases with logger
	phasesToRegister := []phases.Phase{
		phases.NewBootstrapPhase(cfg, log),
		phases.NewPreflightPhase(cfg, log),
		phases.NewPartitioningPhase(cfg, log),
		phases.NewBaseInstallPhase(cfg, log),
		phases.NewConfigPhase(cfg, log),
		phases.NewBootPhase(cfg, log),
		phases.NewReposPhase(cfg, log),
		phases.NewPostInstallPhase(cfg, log),
	}

	for _, phase := range phasesToRegister {
		if err := orchestrator.RegisterPhase(phase); err != nil {
			log.Error("Failed to register phase", "error", err)
			os.Exit(1)
		}
	}

	// Create UI model
	model := ui.NewModel(orchestrator, cfg, log, version)

	// Run Bubbletea app
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("Error running installer", "error", err)
		os.Exit(1)
	}
}
