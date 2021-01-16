package store

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestHeaders_GetSet(t *testing.T) {
	db, done := setupLevelDB(t)

	var expMB blob.SectorHashes
	_, err := rand.Read(expMB[0][:])
	require.NoError(t, err)

	var sig crypto.Signature
	_, err = rand.Read(sig[:])
	require.NoError(t, err)

	expHeader := &Header{
		Name:         "foo",
		EpochHeight:  uint16(0),
		SectorSize:   uint16(0),
		SectorTipHash:   crypto.Rand32(),
		Signature:    sig,
		ReservedRoot: crypto.Rand32(),
		EpochStartAt:   time.Unix(11, 0),
	}
	_, err = GetHeader(db, "foo")
	require.Error(t, err)
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		return SetHeaderTx(tx, expHeader, expMB)
	}))
	actHeader, err := GetHeader(db, "foo")
	require.NoError(t, err)
	require.Equal(t, expHeader.Name, actHeader.Name)
	require.Equal(t, expHeader.EpochHeight, actHeader.EpochHeight)
	require.Equal(t, expHeader.SectorSize, actHeader.SectorSize)
	require.Equal(t, expHeader.SectorTipHash, actHeader.SectorTipHash)
	require.Equal(t, expHeader.Signature, actHeader.Signature)
	require.Equal(t, expHeader.ReservedRoot, actHeader.ReservedRoot)
	require.Equal(t, expHeader.EpochStartAt.Unix(), actHeader.EpochStartAt.Unix())
	actMB, err := GetSectorHashes(db, "foo")
	require.NoError(t, err)
	require.Equal(t, expMB, actMB)

	done()
}
