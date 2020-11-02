package cmd

import (
	"fmt"
	"fnd/cli"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the CLI's home directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cli.InitHomeDir(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully initialized fnd-cli in %s.\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
