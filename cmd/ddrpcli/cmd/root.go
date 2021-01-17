package cmd

import (
	"fmt"
	"fnd/cli"
	"fnd/cmd/ddrpcli/cmd/blob"
	"fnd/cmd/ddrpcli/cmd/net"
	"fnd/cmd/ddrpcli/cmd/unsafe"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ddrpcli",
	Short: "Command-line RPC interface for DDRP.",
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
	rootCmd.PersistentFlags().String(cli.FlagHome, "~/.ddrpcli", "Home directory for the CLI's configuration.")
	rootCmd.PersistentFlags().String(cli.FlagFormat, "text", "Output format")
	net.AddCmd(rootCmd)
	blob.AddCmd(rootCmd)
	unsafe.AddCmd(rootCmd)
}
