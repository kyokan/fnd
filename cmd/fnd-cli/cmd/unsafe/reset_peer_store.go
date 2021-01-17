package unsafe

import (
	"fmt"
	"fnd/config"
	"fnd/store"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var resetPeerStoreCmd = &cobra.Command{
	Use:   "reset-peer-store",
	Short: "Wipes fnd's peer store directly on disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		homePath := config.ExpandHomePath(fndHome)
		db, err := store.Open(config.ExpandDBPath(homePath))
		if err != nil {
			return errors.Wrap(err, "error opening store")
		}
		if err := store.TruncatePeerStore(db); err != nil {
			return errors.Wrap(err, "error truncating peer store")
		}
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "error closing DB")
		}
		fmt.Println("Peer store wiped.")
		return nil
	},
}

func init() {
	resetPeerStoreCmd.Flags().StringVar(&fndHome, "fnd-home", "~/.fnd", "Path to fnd's home directory.")
	cmd.AddCommand(resetPeerStoreCmd)
}
