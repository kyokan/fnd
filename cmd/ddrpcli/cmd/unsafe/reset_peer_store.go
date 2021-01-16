package unsafe

import (
	"fmt"
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var resetPeerStoreCmd = &cobra.Command{
	Use:   "reset-peer-store",
	Short: "Wipes ddrpd's peer store directly on disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		homePath := config.ExpandHomePath(ddrpdHome)
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
	resetPeerStoreCmd.Flags().StringVar(&ddrpdHome, "ddrpd-home", "~/.ddrpd", "Path to DDRPD's home directory.")
	cmd.AddCommand(resetPeerStoreCmd)
}
