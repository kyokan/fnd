package protocol

import (
	"testing"
	"time"

	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/testutil"
	"github.com/ddrp-org/ddrp/testutil/testcrypto"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestUpdateQueue_Enqueue_InvalidBeforeEnqueue(t *testing.T) {
	db, done := setupDB(t)
	defer done()

	identicalHeader := signHeader(t, &store.Header{
		Name:        "identical",
		EpochHeight: uint16(0),
		SectorSize:  uint16(1),
		EpochStartAt:  time.Unix(1, 0),
	})
	throttledHeader := signHeader(t, &store.Header{
		Name:        "throttled",
		EpochHeight: uint16(0),
		SectorSize:  uint16(1),
		EpochStartAt:  time.Now(),
	})
	staleHeader := signHeader(t, &store.Header{
		Name:        "stale",
		EpochHeight: uint16(0),
		SectorSize:  uint16(100),
		EpochStartAt:  time.Unix(1, 0),
	})

	headers := []*store.Header{
		identicalHeader,
		throttledHeader,
		staleHeader,
	}

	_, pub := testcrypto.FixedKey(t)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.SetInitialImportCompleteTx(tx); err != nil {
			return err
		}
		if err := store.BanName(tx, "banned"); err != nil {
			return err
		}
		if err := store.SetNameInfoTx(tx, "banned", pub, 10); err != nil {
			return err
		}
		for _, header := range headers {
			if err := store.SetNameInfoTx(tx, header.Name, pub, 10); err != nil {
				return err
			}
			if err := store.SetHeaderTx(tx, header, blob.ZeroSectorHashes); err != nil {
				return err
			}
		}
		return nil
	}))

	invalid := []struct {
		name     string
		update   *wire.Update
		errCheck func(t *testing.T, err error)
	}{
		{
			"invalid name",
			&wire.Update{
				Name: "--not-a-good-name--",
			},
			func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "name is invalid")
			},
		},
		{
			"banned name",
			&wire.Update{
				Name: "banned",
			},
			func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "name is banned")
			},
		},
		{
			"bad signature",
			&wire.Update{
				Name:          identicalHeader.Name,
				EpochHeight:   identicalHeader.EpochHeight,
				SectorSize:    identicalHeader.SectorSize + 10,
				SectorTipHash: identicalHeader.SectorTipHash,
				ReservedRoot:  identicalHeader.ReservedRoot,
				Signature:     identicalHeader.Signature,
			},
			func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "signature is invalid")
			},
		},
		{
			"identical",
			&wire.Update{
				Name:          identicalHeader.Name,
				EpochHeight:   identicalHeader.EpochHeight,
				SectorSize:    identicalHeader.SectorSize,
				SectorTipHash: identicalHeader.SectorTipHash,
				ReservedRoot:  identicalHeader.ReservedRoot,
				Signature:     identicalHeader.Signature,
			},
			func(t *testing.T, err error) {
				require.Equal(t, ErrUpdateQueueIdenticalTimestamp, err)
			},
		},
		{
			"stale",
			signUpdate(t, &wire.Update{
				Name:          staleHeader.Name,
				EpochHeight:   staleHeader.EpochHeight,
				SectorSize:    staleHeader.SectorSize - 10,
				SectorTipHash: throttledHeader.SectorTipHash,
				ReservedRoot:  identicalHeader.ReservedRoot,
			}),
			func(t *testing.T, err error) {
				require.Equal(t, ErrUpdateQueueStaleTimestamp, err)
			},
		},
	}
	queue := NewUpdateQueue(p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t)), db)
	for _, inv := range invalid {
		t.Run(inv.name, func(t *testing.T) {
			inv.errCheck(t, queue.Enqueue(crypto.Rand32(), inv.update))
		})
	}
}

func TestUpdateQueue_Enqueue_InvalidAfterEnqueue(t *testing.T) {
	db, done := setupDB(t)
	defer done()

	header := signHeader(t, &store.Header{
		Name:        "somename",
		EpochHeight: uint16(0),
		SectorSize:  uint16(100),
		EpochStartAt:  time.Unix(1, 0),
	})

	_, pub := testcrypto.FixedKey(t)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.SetInitialImportCompleteTx(tx); err != nil {
			return err
		}
		if err := store.SetNameInfoTx(tx, header.Name, pub, 10); err != nil {
			return err
		}
		if err := store.SetHeaderTx(tx, header, blob.ZeroSectorHashes); err != nil {
			return err
		}
		return nil
	}))

	queue := NewUpdateQueue(p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t)), db)
	require.NoError(t, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:        header.Name,
		EpochHeight: header.EpochHeight,
		SectorSize:  header.SectorSize + 1,
	})))
	require.Equal(t, ErrUpdateQueueStaleTimestamp, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:        header.Name,
		EpochHeight: header.EpochHeight,
		SectorSize:  header.SectorSize - 10,
	})))
	require.Equal(t, ErrUpdateQueueSpltBrain, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:          header.Name,
		EpochHeight:   header.EpochHeight,
		SectorSize:    header.SectorSize + 1,
		SectorTipHash: crypto.Rand32(),
	})))
}

func TestUpdateQueue_EnqueueDequeue(t *testing.T) {
	db, done := setupDB(t)
	defer done()

	header := signHeader(t, &store.Header{
		Name:        "somename",
		EpochHeight: uint16(0),
		SectorSize:  uint16(100),
		EpochStartAt:  time.Unix(1, 0),
	})

	_, pub := testcrypto.FixedKey(t)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.SetInitialImportCompleteTx(tx); err != nil {
			return err
		}
		if err := store.SetNameInfoTx(tx, header.Name, pub, 10); err != nil {
			return err
		}
		if err := store.SetHeaderTx(tx, header, blob.ZeroSectorHashes); err != nil {
			return err
		}
		return nil
	}))

	pids := []crypto.Hash{
		crypto.Rand32(),
		crypto.Rand32(),
	}
	queue := NewUpdateQueue(p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t)), db)
	update := signUpdate(t, &wire.Update{
		Name:        header.Name,
		EpochHeight: header.EpochHeight,
		SectorSize:  header.SectorSize + 1,
	})
	for _, pid := range pids {
		require.NoError(t, queue.Enqueue(pid, update))
	}

	item := queue.Dequeue()
	for _, pid := range pids {
		require.True(t, item.PeerIDs.Has(pid))
	}
	require.Equal(t, update.Name, item.Name)
	require.Equal(t, update.EpochHeight, item.EpochHeight)
	require.Equal(t, update.SectorSize, item.SectorSize)
	require.Equal(t, update.SectorTipHash, item.SectorTipHash)
	require.Equal(t, update.ReservedRoot, item.ReservedRoot)
	require.Equal(t, update.Signature, item.Signature)
	require.True(t, pub.IsEqual(item.Pub))
	require.Equal(t, 10, item.Height)
}

func signHeader(t *testing.T, header *store.Header) *store.Header {
	sig, err := blob.SignSeal(testcrypto.FixedSigner(t), header.Name, header.EpochHeight, header.SectorSize, header.SectorTipHash, header.ReservedRoot)
	require.NoError(t, err)
	header.Signature = sig
	return header
}

func signUpdate(t *testing.T, update *wire.Update) *wire.Update {
	sig, err := blob.SignSeal(testcrypto.FixedSigner(t), update.Name, update.EpochHeight, update.SectorSize, update.SectorTipHash, update.ReservedRoot)
	require.NoError(t, err)
	update.Signature = sig
	return update
}
