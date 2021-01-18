package blob

import (
	"encoding/json"
	"fnd/cli"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"fnd/store"
	"github.com/spf13/cobra"
	"math"
	"os"
	"strconv"
)

var listCmd = &cobra.Command{
	Use:   "list <start?> <limit?>",
	Short: "Lists blobs.",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var start string
		if len(args) >= 1 {
			start = args[0]
		}
		lim := math.MaxInt64
		if len(args) == 2 {
			limit, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return err
			}
			lim = int(limit)
		}

		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		grpcClient := apiv1.NewFootnotev1Client(conn)
		var count int
		encoder := json.NewEncoder(os.Stdout)
		var innerErr error
		err = rpc.ListBlobInfo(grpcClient, start, func(info *store.BlobInfo) bool {
			if err := encoder.Encode(info); err != nil {
				innerErr = err
				return false
			}
			count++
			return count < lim
		})
		if err != nil {
			return err
		}
		if innerErr != nil {
			return err
		}
		return nil
	},
}

func init() {
	cmd.AddCommand(listCmd)
}
