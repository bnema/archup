package cmd

import (
	"fmt"
	"os"

	apphandlers "github.com/bnema/archup/internal/application/handlers"
	"github.com/bnema/archup/internal/application/services"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/infrastructure/executor"
	"github.com/bnema/archup/internal/infrastructure/filesystem"
	infrahttp "github.com/bnema/archup/internal/infrastructure/http"
	infralogger "github.com/bnema/archup/internal/infrastructure/logger"
	"github.com/bnema/archup/internal/infrastructure/persistence"
	"github.com/bnema/archup/internal/interfaces/tui"
	"github.com/bnema/archup/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newInstallCmd())
}

func newInstallCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Run base system installer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show TUI but don't execute commands")
	return cmd
}

func runInstall(dryRun bool) error {
	oldLog, err := logger.New(config.DefaultLogPath, dryRun)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() {
		if err := oldLog.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
		}
	}()

	if dryRun {
		oldLog.Info("Running in DRY-RUN mode - commands will be logged but not executed")
	}

	cfg := config.NewConfig(version)
	oldLog.Info("Config initialized", "version", version, "raw_url", cfg.RawURL)

	oldLog.Info("Initializing DDD architecture components")

	slogAdapter := infralogger.NewSlogAdapter(oldLog.Slog())

	fsAdapter := &filesystem.LocalFileSystem{}
	shellExec := executor.NewShellExecutor(slogAdapter)
	chrootExec := executor.NewChrootExecutor(slogAdapter)
	httpClient := infrahttp.NewHTTPClient()
	scriptExec := executor.NewScriptExecutor(fsAdapter, shellExec, config.DefaultInstallDir)
	repoAdapter, err := persistence.NewFileRepository(config.DefaultConfigPath)
	if err != nil {
		oldLog.Error("Failed to create repository adapter", "error", err)
		return fmt.Errorf("create repository adapter: %w", err)
	}

	bootstrapHandler := apphandlers.NewBootstrapHandler(fsAdapter, httpClient, slogAdapter, cfg.RepoURL, cfg.RawURL)
	preflightHandler := apphandlers.NewPreflightHandler(fsAdapter, shellExec, slogAdapter)
	partitionHandler := apphandlers.NewPartitionHandler(shellExec, slogAdapter)
	baseHandler := apphandlers.NewInstallBaseHandler(fsAdapter, shellExec, chrootExec, slogAdapter)
	configHandler := apphandlers.NewConfigureSystemHandler(fsAdapter, chrootExec, slogAdapter)
	bootloaderHandler := apphandlers.NewBootloaderHandler(fsAdapter, shellExec, chrootExec, slogAdapter)
	reposHandler := apphandlers.NewReposHandler(fsAdapter, chrootExec, slogAdapter)
	postInstallHandler := apphandlers.NewPostInstallHandler(fsAdapter, httpClient, chrootExec, scriptExec, slogAdapter, cfg.RawURL)

	installService := services.NewInstallationService(
		repoAdapter,
		slogAdapter,
		bootstrapHandler,
		preflightHandler,
		partitionHandler,
		baseHandler,
		configHandler,
		bootloaderHandler,
		reposHandler,
		postInstallHandler,
	)

	gpuHandler := apphandlers.NewGPUHandler(shellExec, slogAdapter)
	tuiApp := tui.NewApp(installService, installService.Tracker(), gpuHandler, slogAdapter)

	oldLog.Info("Starting TUI application", "version", version)
	p := tea.NewProgram(tuiApp, tea.WithAltScreen())
	tuiApp.SetProgram(p)
	if _, err := p.Run(); err != nil {
		oldLog.Error("Error running installer", "error", err)
		return fmt.Errorf("run installer: %w", err)
	}

	if err := tuiApp.Close(); err != nil {
		oldLog.Error("Error closing TUI app", "error", err)
	}

	return nil
}
