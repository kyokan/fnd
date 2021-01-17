package blob

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"

	"github.com/mslipper/handshake/primitives"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <names>",
	Short: "Returns metadata about DDRP blobs.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		names := strings.Split(args[0], ",")
		for _, name := range names {
			if err := primitives.ValidateName(name); err != nil {
				return errors.Wrap(err, fmt.Sprintf("invalid name %s", name))
			}
		}

		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewDDRPv1Client(conn)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Name",
			"Public Key",
			"Timestamp",
			"Merkle Root",
			"Reserved Root",
			"Received At",
			"Signature",
			"Time Bank",
		})

		for _, name := range names {
			res, err := rpc.GetBlobInfo(grpcClient, name)
			if err != nil {
				return err
			}

			table.Append([]string{
				res.Name,
				hex.EncodeToString(res.PublicKey.SerializeCompressed()),
				strconv.Itoa(int(res.EpochHeight)),
				strconv.Itoa(int(res.SectorSize)),
				res.MerkleRoot.String(),
				res.ReservedRoot.String(),
				res.ReceivedAt.String(),
				res.Signature.String(),
			})
		}

		table.Render()
		return nil
	},
}

func init() {
	cmd.AddCommand(infoCmd)
}
