package main

import (
	"flag"
	"fmt"
	"os"

	apphandlers "github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/application/services"
	"github.com/bnema/archup/internal/cleanup"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/infrastructure/executor"
	"github.com/bnema/archup/internal/infrastructure/filesystem"
	infralogger "github.com/bnema/archup/internal/infrastructure/logger"
	"github.com/bnema/archup/internal/infrastructure/persistence"
	"github.com/bnema/archup/internal/interfaces/tui"
	"github.com/bnema/archup/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
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
	oldLog, err := logger.New(config.DefaultLogPath, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer oldLog.Close()

	// Handle --cleanup flag
	if *doCleanup {
		if err := cleanup.Run(oldLog); err != nil {
			oldLog.Error("Cleanup failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Log dry-run mode if enabled
	if *dryRun {
		oldLog.Info("Running in DRY-RUN mode - commands will be logged but not executed")
	}

	// Load or create config (pass version to determine correct branch for downloads)
	cfg := config.NewConfig(version)
	oldLog.Info("Config initialized", "version", version, "raw_url", cfg.RawURL)

	// ==================== DEPENDENCY INJECTION ====================
	oldLog.Info("Initializing DDD architecture components")

	// Adapt old logger to ports.Logger interface
	slogAdapter := infralogger.NewSlogAdapter(oldLog.Slog())

	// 1. Create infrastructure adapters
	fsAdapter := &filesystem.LocalFileSystem{}
	shellExec := executor.NewShellExecutor(slogAdapter)
	chrootExec := executor.NewChrootExecutor(slogAdapter)
	scriptExec := executor.NewScriptExecutor(fsAdapter, shellExec, config.DefaultInstallDir)
	repoAdapter, err := persistence.NewFileRepository(config.DefaultConfigPath)
	if err != nil {
		oldLog.Error("Failed to create repository adapter", "error", err)
		os.Exit(1)
	}

	// 2. Create command handlers (inject ports)
	preflightHandler := apphandlers.NewPreflightHandler(fsAdapter, shellExec, slogAdapter)
	partitionHandler := apphandlers.NewPartitionHandler(shellExec, slogAdapter)
	baseHandler := apphandlers.NewInstallBaseHandler(fsAdapter, shellExec, chrootExec, slogAdapter)
	configHandler := apphandlers.NewConfigureSystemHandler(chrootExec, slogAdapter)
	bootloaderHandler := apphandlers.NewBootloaderHandler(chrootExec, slogAdapter)
	reposHandler := apphandlers.NewReposHandler(chrootExec, slogAdapter)
	postInstallHandler := apphandlers.NewPostInstallHandler(chrootExec, scriptExec, slogAdapter)

	// 3. Create application services
	installService := services.NewInstallationService(
		repoAdapter,
		slogAdapter,
		preflightHandler,
		partitionHandler,
		baseHandler,
		configHandler,
		bootloaderHandler,
		reposHandler,
		postInstallHandler,
	)

	// 4. Create TUI app with services
	tuiApp := tui.NewApp(installService, installService.Tracker(), slogAdapter)

	// ==================== RUN BUBBLETEA ====================
	oldLog.Info("Starting TUI application", "version", version)
	p := tea.NewProgram(tuiApp, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		oldLog.Error("Error running installer", "error", err)
		os.Exit(1)
	}

	// Cleanup
	if err := tuiApp.Close(); err != nil {
		oldLog.Error("Error closing TUI app", "error", err)
	}
}
