package net

import (
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Returns network status information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewDDRPv1Client(conn)

		res, err := rpc.GetStatus(grpcClient)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Append([]string{
			"Peer ID", res.PeerID,
		})
		table.Append([]string{
			"Peer Count", strconv.Itoa(res.PeerCount),
		})
		table.Append([]string{
			"Header Count", strconv.Itoa(res.HeaderCount),
		})
		table.Append([]string{
			"Tx Bytes",
			bandwidthToStr(res.TxBytes),
		})
		table.Append([]string{
			"Rx Bytes",
			bandwidthToStr(res.RxBytes),
		})
		table.Render()
		return nil
	},
}

func init() {
	cmd.AddCommand(statusCmd)
}
