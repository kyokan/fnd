package store

import (
	"fnd/testutil/testcrypto"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestNaming_Meta(t *testing.T) {
	db, done := setupLevelDB(t)

	height, err := GetLastNameImportHeight(db)
	require.NoError(t, err)
	require.Equal(t, 0, height)

	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return SetLastNameImportHeightTx(tx, 100)
	}))
	height, err = GetLastNameImportHeight(db)
	require.NoError(t, err)
	require.Equal(t, 100, height)

	done()
}

func TestNaming_GetSetNameInfo(t *testing.T) {
	db, done := setupLevelDB(t)

	info, err := GetNameInfo(db, "foo")
	require.Error(t, err)

	_, pub := testcrypto.RandKey()
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return SetNameInfoTx(tx, "foo", pub, 10)
	}))
	info, err = GetNameInfo(db, "foo")
	require.NoError(t, err)
	require.EqualValues(t, &NameInfo{
		Name:         "foo",
		PublicKey:    pub,
		ImportHeight: 10,
	}, info)

	done()
}

func TestNaming_StreamNameInfo(t *testing.T) {
	db, done := setupLevelDB(t)

	_, pub := testcrypto.RandKey()
	items := []*NameInfo{
		{
			Name:         "bar",
			PublicKey:    pub,
			ImportHeight: 11,
		},
		{
			Name:         "baz",
			PublicKey:    pub,
			ImportHeight: 12,
		},
		{
			Name:         "foo",
			PublicKey:    pub,
			ImportHeight: 10,
		},
	}
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		for _, item := range items {
			require.NoError(t, SetNameInfoTx(tx, item.Name, item.PublicKey, item.ImportHeight))
		}
		return nil
	}))
	stream, err := StreamNameInfo(db, "")
	require.NoError(t, err)
	var i int
	for {
		item, err := stream.Next()
		require.NoError(t, err)
		if item == nil {
			break
		}
		require.EqualValues(t, items[i], item)
		i++
	}
	require.Equal(t, 3, i)
	require.NoError(t, stream.Close())

	stream, err = StreamNameInfo(db, "baz")
	require.NoError(t, err)
	item, err := stream.Next()
	require.NoError(t, err)
	require.EqualValues(t, items[2], item)
	item, err = stream.Next()
	require.Nil(t, item)
	require.NoError(t, err)

	done()
}
