package unsafe

import (
	"fmt"
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var resetNameStore = &cobra.Command{
	Use:   "reset-name-store",
	Short: "Wipes ddrpd's naming data directly on disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		homePath := config.ExpandHomePath(ddrpdHome)
		db, err := store.Open(config.ExpandDBPath(homePath))
		if err != nil {
			return errors.Wrap(err, "failed to open store")
		}
		if err := store.TruncateNameStore(db); err != nil {
			return errors.Wrap(err, "error truncating name store")
		}
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "error closing DB")
		}
		fmt.Println("HNS name store wiped.")
		return nil
	},
}

func init() {
	resetNameStore.Flags().StringVar(&ddrpdHome, "ddrpd-home", "~/.ddrpd", "Path to DDRPD's home directory.")
	cmd.AddCommand(resetNameStore)
}
