package net

import (
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"

	"github.com/spf13/cobra"
)

var unbanPeerCmd = &cobra.Command{
	Use:   "unban-peer <ip>",
	Short: "Unbans a peer.",
	Long: `Unbans a peer. A connection with the peer will not be automatically reestablished;
fnd will either reconnect to the unbanned peer the next time it refills its 
peer list or following the add-peer CLI command.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewFootnotev1Client(conn)
		return rpc.UnbanPeer(grpcClient, args[0])
	},
}

func init() {
	cmd.AddCommand(unbanPeerCmd)
}
