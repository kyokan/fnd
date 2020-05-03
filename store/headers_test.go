package store

import (
	"crypto/rand"
	"ddrp/blob"
	"ddrp/crypto"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
)

func TestHeaders_GetSet(t *testing.T) {
	db, done := setupLevelDB(t)

	var expMB blob.MerkleBase
	_, err := rand.Read(expMB[0][:])
	require.NoError(t, err)

	var sig crypto.Signature
	_, err = rand.Read(sig[:])
	require.NoError(t, err)

	expHeader := &Header{
		Name:         "foo",
		Timestamp:    time.Unix(10, 0),
		MerkleRoot:   crypto.Rand32(),
		Signature:    sig,
		ReservedRoot: crypto.Rand32(),
		ReceivedAt:   time.Unix(11, 0),
		Timebank:     100,
	}
	_, err = GetHeader(db, "foo")
	require.Error(t, err)
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return SetHeaderTx(tx, expHeader, expMB)
	}))
	actHeader, err := GetHeader(db, "foo")
	require.NoError(t, err)
	require.Equal(t, expHeader.Name, actHeader.Name)
	require.Equal(t, expHeader.Timestamp.Unix(), actHeader.Timestamp.Unix())
	require.Equal(t, expHeader.MerkleRoot, actHeader.MerkleRoot)
	require.Equal(t, expHeader.Signature, actHeader.Signature)
	require.Equal(t, expHeader.ReservedRoot, actHeader.ReservedRoot)
	require.Equal(t, expHeader.ReceivedAt.Unix(), actHeader.ReceivedAt.Unix())
	actMB, err := GetMerkleBase(db, "foo")
	require.NoError(t, err)
	require.Equal(t, expMB, actMB)

	done()
}
