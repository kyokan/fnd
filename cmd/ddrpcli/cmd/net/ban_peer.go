package net

import (
	"github.com/ddrp-org/ddrp/cli"
	"github.com/ddrp-org/ddrp/rpc"
	apiv1 "github.com/ddrp-org/ddrp/rpc/v1"
	"github.com/spf13/cobra"
	"strconv"
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
		grpcClient := apiv1.NewDDRPv1Client(conn)
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
