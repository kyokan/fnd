package cmd

import (
	"ddrp/cli"
	"fmt"
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

		fmt.Printf("Successfully initialized ddrpcli in %s.\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
