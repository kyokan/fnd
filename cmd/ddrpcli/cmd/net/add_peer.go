package net

import (
	"ddrp/cli"
	"ddrp/rpc"
	apiv1 "ddrp/rpc/v1"
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

var addPeerCmd = &cobra.Command{
	Use:   "add-peer <peer-id?>@<ip> <?verify>",
	Short: "Adds a peer.",
	Long:  `Adds a peer. If the peer is banned, this command is a no-op.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewDDRPv1Client(conn)

		verify := len(args) > 1 && args[1] == "true"

		splits := strings.Split(args[0], "@")
		switch len(splits) {
		case 1:
			return rpc.AddPeer(grpcClient, "", splits[0], verify)
		case 2:
			return rpc.AddPeer(grpcClient, splits[0], splits[1], verify)
		default:
			return errors.New("must specify the peer as <peer-id?>@<ip>")
		}
	},
}

func init() {
	cmd.AddCommand(addPeerCmd)
}
