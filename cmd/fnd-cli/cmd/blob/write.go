package blob

import (
	"bufio"
	"bytes"
	"fmt"
	"fnd/blob"
	"fnd/cli"
	"fnd/config"
	"fnd/protocol"
	"fnd/rpc"
	apiv1 "fnd/rpc/v1"
	"fnd/store"
	"io"
	"os"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	BroadcastFlag  = "broadcast"
	ResetEpochFlag = "reset-epoch"
)

var (
	fndHome    string
	broadcast  bool
	resetEpoch bool
)

var writeCmd = &cobra.Command{
	Use:   "write <name> <data>",
	Short: "Write data to the specified blob.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		conn, err := cli.DialRPC(cmd)
		if err != nil {
			return err
		}

		if resetEpoch {
			res, err := rpc.GetBlobInfo(apiv1.NewFootnotev1Client(conn), name)
			if err != nil {
				return err
			}

			if res.EpochHeight >= protocol.CurrentEpoch(name) {
				return errors.New("cannot reset epoch ahead of schedule")
			}

			homePath := config.ExpandHomePath(fndHome)
			db, err := store.Open(config.ExpandDBPath(homePath))
			if err != nil {
				return errors.Wrap(err, "error opening store")
			}

			blobsPath := config.ExpandBlobsPath(homePath)
			blobSubpath := blob.PathifyName(name)
			blobFile := path.Join(blobsPath, blobSubpath)
			if err := os.RemoveAll(blobFile); err != nil {
				return errors.Wrap(err, "error erasing blob data")
			}
			if err := store.TruncateHeaderName(db, name); err != nil {
				return errors.Wrap(err, "error truncating header store")
			}
			if err := db.Close(); err != nil {
				return errors.Wrap(err, "error closing DB")
			}
		}

		homeDir := cli.GetHomeDir(cmd)
		signer, err := cli.GetSigner(homeDir)
		if err != nil {
			return err
		}

		wr := rpc.NewBlobWriter(apiv1.NewFootnotev1Client(conn), signer, name)

		if err := wr.Open(); err != nil {
			return err
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
		var sector blob.Sector
		for i := 0; i < blob.SectorCount; i++ {
			if _, err := rd.Read(sector[:]); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			wr.WriteSector(sector[:])
		}
		sectorTipHash, err := wr.Commit(broadcast)
		if err != nil {
			return err
		}

		fmt.Printf("Success. Hash: %v\n", sectorTipHash)
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
	writeCmd.Flags().BoolVar(&broadcast, BroadcastFlag, true, "Broadcast data to the network upon completion")
	writeCmd.Flags().BoolVar(&resetEpoch, ResetEpochFlag, false, "Increment the epoch and reset the blob before write.")
	writeCmd.Flags().StringVar(&fndHome, "fnd-home", "~/.fnd", "Path to FootnoteD's home directory.")
	cmd.AddCommand(writeCmd)
}
