package net

import (
	"errors"
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"github.com/spf13/cobra"
	"strings"
)

var (
	verifyPeerID bool
)

var addPeerCmd = &cobra.Command{
	Use:   "add-peer <peer-id?>@<ip>",
	Short: "Adds a peer.",
	Long:  `Adds a peer. If the peer is banned, this command is a no-op.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewFootnotev1Client(conn)
		splits := strings.Split(args[0], "@")
		if verifyPeerID && len(splits) == 1 {
			return errors.New("must define a peer ID if the peer ID is to be verified")
		}

		switch len(splits) {
		case 1:
			return rpc.AddPeer(grpcClient, "", splits[0], verifyPeerID)
		case 2:
			return rpc.AddPeer(grpcClient, splits[0], splits[1], verifyPeerID)
		default:
			return errors.New("must specify the peer as <peer-id?>@<ip>")
		}
	},
}

func init() {
	cmd.AddCommand(addPeerCmd)
	addPeerCmd.Flags().BoolVar(&verifyPeerID, "verify", false, "Verify the remote peer's ID")
}
