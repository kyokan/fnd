package cmd

import (
	"fmt"
	"fnd/cli"
	"fnd/config"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var configuredHomeDir string

var rootCmd = &cobra.Command{
	Use:   "fnd",
	Short: "fnd Daemon",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == "init" {
			return nil
		}
		configuredHomeDir = cli.GetHomeDir(cmd)
		if err := config.EnsureHomeDir(configuredHomeDir); err != nil {
			return errors.Wrap(err, "error ensuring home directory")
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().String(cli.FlagHome, "~/.fnd", "Home directory for the daemon's config and database.")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
