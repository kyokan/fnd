package blob

import (
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <name>",
	Short: "Reads data from the specified blob.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}

		br := rpc.NewBlobReader(apiv1.NewDDRPv1Client(conn), name)
		if _, err := io.Copy(os.Stdout, br); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	cmd.AddCommand(readCmd)
}
