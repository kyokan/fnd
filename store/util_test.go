package store

import (
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"os"
	"testing"
)

func setupLevelDB(t *testing.T) (*leveldb.DB, func()) {
	tmp, err := ioutil.TempDir("", "testdb_")
	require.NoError(t, err)
	db, err := leveldb.OpenFile(tmp, nil)
	require.NoError(t, err)

	return db, func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(tmp))
	}
}
