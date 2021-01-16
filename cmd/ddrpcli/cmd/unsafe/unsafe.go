package unsafe

import "github.com/spf13/cobra"

var ddrpdHome string

var cmd = &cobra.Command{
	Use:   "unsafe",
	Short: "Commands that have dangerous side-effects. Used during development or debugging.",
}

func AddCmd(parent *cobra.Command) {
	parent.AddCommand(cmd)
}
