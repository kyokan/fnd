package net

import (
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"strconv"

	"github.com/spf13/cobra"
)

var banPeerCmd = &cobra.Command{
	Use:   "ban-peer <peer-id> <duration-ms>",
	Short: "Bans a peer for the given duration.",
	Long: `Bans a peer for the given duration in milliseconds. Any existing connections
to this peer will be closed.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewFootnotev1Client(conn)
		peerID := args[0]
		duration, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		return rpc.BanPeer(grpcClient, peerID, duration)
	},
}

func init() {
	cmd.AddCommand(banPeerCmd)
}
