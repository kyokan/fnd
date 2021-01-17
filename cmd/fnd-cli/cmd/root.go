package cmd

import (
	"fmt"
	"fnd/cli"
	"fnd/cmd/fnd-cli/cmd/blob"
	"fnd/cmd/fnd-cli/cmd/net"
	"fnd/cmd/fnd-cli/cmd/unsafe"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fnd-cli",
	Short: "Command-line RPC interface for fnd.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Int(cli.FlagRPCPort, 9098, "RPC port to connect to.")
	rootCmd.PersistentFlags().String(cli.FlagRPCHost, "127.0.0.1", "RPC host to connect to.")
	rootCmd.PersistentFlags().String(cli.FlagHome, "~/.fnd-cli", "Home directory for the CLI's configuration.")
	rootCmd.PersistentFlags().String(cli.FlagFormat, "text", "Output format")
	net.AddCmd(rootCmd)
	blob.AddCmd(rootCmd)
	unsafe.AddCmd(rootCmd)
}
