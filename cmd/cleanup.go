package cmd

import (
	"fmt"
	"os"

	"github.com/bnema/archup/internal/cleanup"
	"github.com/bnema/archup/internal/config"
	"github.com/bnema/archup/internal/logger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCleanupCmd())
}

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up installation artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			oldLog, err := logger.New(config.DefaultLogPath, false)
			if err != nil {
				return fmt.Errorf("failed to create logger: %w", err)
			}
			defer func() {
				if err := oldLog.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
				}
			}()

			if err := cleanup.Run(oldLog); err != nil {
				oldLog.Error("Cleanup failed", "error", err)
				return fmt.Errorf("cleanup failed: %w", err)
			}

			return nil
		},
	}
}
