package net

import (
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "net",
	Short: "Commands related to fnd's network connection.",
}

func AddCmd(parent *cobra.Command) {
	parent.AddCommand(cmd)
}
