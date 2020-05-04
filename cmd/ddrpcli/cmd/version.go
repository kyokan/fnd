package cmd

import (
	"fmt"
	"github.com/ddrp-org/ddrp/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use: "version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("ddrpcli %s (%s)\n", version.GitTag, version.GitCommit)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
