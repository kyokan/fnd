package blob

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ddrp-org/ddrp/cli"
	"github.com/ddrp-org/ddrp/rpc"
	apiv1 "github.com/ddrp-org/ddrp/rpc/v1"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"io"
	"os"
)

const (
	TruncateFlag  = "truncate"
	BroadcastFlag = "broadcast"
)

var (
	truncate  bool
	broadcast bool
)

var writeCmd = &cobra.Command{
	Use:   "write <name> <data>",
	Short: "Write data to the specified blob.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}
		homeDir := cli.GetHomeDir(cmd)
		signer, err := cli.GetSigner(homeDir)
		if err != nil {
			return err
		}

		name := args[0]
		wr := rpc.NewBlobWriter(apiv1.NewDDRPv1Client(conn), signer, name)

		if err := wr.Open(); err != nil {
			return err
		}
		if truncate {
			if err := wr.Truncate(); err != nil {
				return err
			}
		}
		var rd io.Reader
		if len(args) < 2 {
			if isatty.IsTerminal(os.Stdin.Fd()) {
				rd = bufio.NewReader(bytes.NewReader(readDataTTY()))
			} else {
				rd = os.Stdin
			}
		} else {
			rd = bufio.NewReader(bytes.NewReader([]byte(args[1])))
		}
		if _, err := io.Copy(wr, rd); err != nil {
			return err
		}
		if err := wr.Commit(broadcast); err != nil {
			return err
		}

		fmt.Println("Success.")
		return nil
	},
}

func readDataTTY() []byte {
	fmt.Println("Paste or type the contents you would like to write below.")
	fmt.Println("When you are finished, press Ctrl+D.")

	var buf bytes.Buffer
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := scanner.Text()
		buf.Write([]byte(data))
		buf.WriteByte('\n')
	}

	return buf.Bytes()
}

func init() {
	writeCmd.Flags().BoolVar(&truncate, TruncateFlag, false, "Truncate the blob before writing")
	writeCmd.Flags().BoolVar(&broadcast, BroadcastFlag, true, "Broadcast data to the network upon completion")
	cmd.AddCommand(writeCmd)
}
