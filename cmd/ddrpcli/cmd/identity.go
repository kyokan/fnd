package cmd

import (
	"encoding/hex"
	"fmt"
	"fnd/cli"
	"fnd/config"

	"github.com/spf13/cobra"
)

var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Prints the CLI's public key.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		homeDir := cli.GetHomeDir(cmd)
		return config.EnsureHomeDir(homeDir)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir := cli.GetHomeDir(cmd)
		signer, err := cli.GetSigner(homeDir)
		if err != nil {
			return err
		}
		pub := signer.Pub()

		fmt.Println(hex.EncodeToString(pub.SerializeCompressed()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(identityCmd)
}
