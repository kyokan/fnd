package store

import (
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
)

func TestModeration_Meta(t *testing.T) {
	db, done := setupLevelDB(t)

	lastImport, err := GetLastBanListImportAt(db)
	require.NoError(t, err)
	require.EqualValues(t, 0, lastImport.Unix())

	now := time.Now()
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return SetLastBanListImportAt(tx, now)
	}))
	lastImport, err = GetLastBanListImportAt(db)
	require.NoError(t, err)
	require.Equal(t, now.Unix(), lastImport.Unix())

	done()
}

func TestModeration_BanLists(t *testing.T) {
	db, done := setupLevelDB(t)

	isBanned, err := NameIsBanned(db, "foo")
	require.NoError(t, err)
	require.False(t, isBanned)

	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		if err := BanName(tx, "foo"); err != nil {
			return err
		}
		return BanName(tx, "bar")
	}))

	isBanned, err = NameIsBanned(db, "foo")
	require.NoError(t, err)
	require.True(t, isBanned)
	isBanned, err = NameIsBanned(db, "bar")
	require.NoError(t, err)
	require.True(t, isBanned)

	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return TruncateBannedNames(tx)
	}))
	isBanned, err = NameIsBanned(db, "foo")
	require.NoError(t, err)
	require.False(t, isBanned)
	isBanned, err = NameIsBanned(db, "bar")
	require.NoError(t, err)
	require.False(t, isBanned)

	done()
}
