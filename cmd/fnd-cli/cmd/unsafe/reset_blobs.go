package unsafe

import (
	"fmt"
	"fnd/config"
	"fnd/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

var resetBlobsCmd = &cobra.Command{
	Use:   "reset-blobs",
	Short: "Wipes fnd's blob data directly on disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		homePath := config.ExpandHomePath(fndHome)
		db, err := store.Open(config.ExpandDBPath(homePath))
		if err != nil {
			return errors.Wrap(err, "error opening store")
		}

		if err := os.RemoveAll(config.ExpandBlobsPath(homePath)); err != nil {
			return errors.Wrap(err, "error erasing blob data")
		}
		if err := config.InitBlobsDir(homePath); err != nil {
			return errors.Wrap(err, "error recreating blobs directory")
		}
		if err := store.TruncateHeaderStore(db); err != nil {
			return errors.Wrap(err, "error truncating header store")
		}
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "error closing DB")
		}
		fmt.Println("Blob data wiped.")
		return nil
	},
}

func init() {
	resetBlobsCmd.Flags().StringVar(&fndHome, "fnd-home", "~/.fnd", "Path to FootnoteD's home directory.")
	cmd.AddCommand(resetBlobsCmd)
}
