package cmd

import (
	"fmt"
	"os"

	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/infrastructure/executor"
	"github.com/bnema/archup/internal/infrastructure/filesystem"
	infralogger "github.com/bnema/archup/internal/infrastructure/logger"
	"github.com/bnema/archup/internal/logger"
	"github.com/bnema/archup/internal/wizard/application/services"
	"github.com/bnema/archup/internal/wizard/interfaces/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newWizardCmd())
}

func newWizardCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Run post-boot desktop wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWizard(dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show TUI but don't execute commands")
	return cmd
}

func runWizard(dryRun bool) error {
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
		oldLog.Info("Running wizard in DRY-RUN mode - commands will be logged but not executed")
	}

	slogAdapter := infralogger.NewSlogAdapter(oldLog.Slog())
	fsAdapter := &filesystem.LocalFileSystem{}
	shellExec := executor.NewShellExecutor(slogAdapter)

	wizardService := services.NewWizardService(fsAdapter, shellExec, slogAdapter)
	app := tui.NewApp(wizardService, slogAdapter)

	oldLog.Info("Starting wizard TUI")
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		oldLog.Error("Error running wizard", "error", err)
		return fmt.Errorf("run wizard: %w", err)
	}

	return nil
}
