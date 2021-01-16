package mockapp

import (
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/testutil/testfs"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

type TestStorage struct {
	BlobStore blob.Store
	DB        *leveldb.DB
}

func CreateStorage(t *testing.T) (*TestStorage, func()) {
	bs, bsDone := CreateTestBlobStore(t)
	db, dbDone := CreateTestDB(t)
	done := func() {
		bsDone()
		dbDone()
	}
	return &TestStorage{
		BlobStore: bs,
		DB:        db,
	}, done
}

func CreateTestBlobStore(t *testing.T) (blob.Store, func()) {
	blobsPath, done := testfs.NewTempDir(t)
	bs := blob.NewStore(blobsPath)
	return bs, done
}

func CreateTestDB(t *testing.T) (*leveldb.DB, func()) {
	dbDir, done := testfs.NewTempDir(t)
	db, err := store.Open(dbDir)
	require.NoError(t, err)
	return db, done
}

func FillBlobReader(t *testing.T, db *leveldb.DB, bs blob.Store, signer crypto.Signer, name string, epochHeight, sectorSize uint16, receivedAt time.Time, r io.Reader) *wire.Update {
	bl, err := bs.Open(name)
	require.NoError(t, err)
	tx, err := bl.Transaction()
	require.NoError(t, err)
	_, err = io.Copy(blob.NewWriter(tx), io.LimitReader(r, blob.Size))
	require.NoError(t, err)
	tree, err := blob.SerialHash(blob.NewReader(tx), blob.ZeroHash, sectorSize)
	require.NoError(t, err)
	sig, err := blob.SignSeal(signer, name, epochHeight, sectorSize, tree.Tip(), crypto.ZeroHash)
	require.NoError(t, err)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		return store.SetHeaderTx(tx, &store.Header{
			Name:          name,
			EpochHeight:   epochHeight,
			SectorSize:    sectorSize,
			SectorTipHash: tree.Tip(),
			Signature:     sig,
			ReservedRoot:  crypto.ZeroHash,
			EpochStartAt:  receivedAt,
		}, tree)
	}))
	require.NoError(t, tx.Commit())
	return &wire.Update{
		Name:          name,
		EpochHeight:   epochHeight,
		SectorSize:    sectorSize,
		SectorTipHash: tree.Tip(),
		ReservedRoot:  crypto.ZeroHash,
		Signature:     sig,
	}
}

func FillBlobRandom(t *testing.T, db *leveldb.DB, bs blob.Store, signer crypto.Signer, name string, epochHeight, sectorSize uint16, receivedAt time.Time) *wire.Update {
	return FillBlobReader(
		t,
		db,
		bs,
		signer,
		name,
		epochHeight,
		sectorSize,
		receivedAt,
		rand.Reader,
	)
}

func RequireBlobsEqual(t *testing.T, localBS blob.Store, remoteBS blob.Store, name string) {
	localBl, err := localBS.Open(name)
	require.NoError(t, err)
	remoteBl, err := remoteBS.Open(name)
	require.NoError(t, err)
	for i := 0; i < blob.SectorCount; i++ {
		localSector, err := localBl.ReadSector(uint8(i))
		require.NoError(t, err)
		remoteSector, err := remoteBl.ReadSector(uint8(i))
		require.NoError(t, err)
		require.EqualValues(t, localSector, remoteSector)
	}
}
