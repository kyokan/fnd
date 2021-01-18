package cmd

import (
	"fmt"
	"fnd/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use: "version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("fnd %s (%s)\n", version.GitTag, version.GitCommit)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
