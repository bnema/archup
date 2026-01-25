package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set via ldflags at build time
// Example: go build -ldflags "-X github.com/bnema/archup/cmd.version=v0.3.0"
var version = "dev"

var rootCmd = &cobra.Command{
	Use:          "archup",
	Short:        "ArchUp installer",
	Long:         "ArchUp installs and configures Arch Linux systems.",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
