package protocol

import (
	"fmt"
	"fnd/store"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
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
