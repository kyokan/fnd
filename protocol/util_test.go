package protocol

import (
	"fmt"
	"github.com/ddrp-org/ddrp/store"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"os"
	"testing"
)

func setupDB(t *testing.T) (*leveldb.DB, func()) {
	tmp, err := ioutil.TempDir("", "testdb_")
	require.NoError(t, err)
	fmt.Println(tmp)
	db, err := store.Open(tmp)
	require.NoError(t, err)

	return db, func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(tmp))
	}
}
