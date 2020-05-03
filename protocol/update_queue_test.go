package protocol

import (
	"ddrp/blob"
	"ddrp/crypto"
	"ddrp/p2p"
	"ddrp/store"
	"ddrp/testutil"
	"ddrp/testutil/testcrypto"
	"ddrp/wire"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
)

func TestUpdateQueue_Enqueue_InvalidBeforeEnqueue(t *testing.T) {
	db, done := setupDB(t)
	defer done()

	identicalHeader := signHeader(t, &store.Header{
		Name:       "identical",
		Timestamp:  time.Unix(1, 0),
		ReceivedAt: time.Unix(1, 0),
	})
	throttledHeader := signHeader(t, &store.Header{
		Name:       "throttled",
		Timestamp:  time.Unix(1, 0),
		ReceivedAt: time.Now(),
	})
	staleHeader := signHeader(t, &store.Header{
		Name:       "stale",
		Timestamp:  time.Unix(100, 0),
		ReceivedAt: time.Unix(1, 0),
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
			if err := store.SetHeaderTx(tx, header, blob.ZeroMerkleBase); err != nil {
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
				Name:         identicalHeader.Name,
				Timestamp:    identicalHeader.Timestamp.Add(10 * time.Second),
				MerkleRoot:   identicalHeader.MerkleRoot,
				ReservedRoot: identicalHeader.ReservedRoot,
				Signature:    identicalHeader.Signature,
			},
			func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "signature is invalid")
			},
		},
		{
			"identical",
			&wire.Update{
				Name:         identicalHeader.Name,
				Timestamp:    identicalHeader.Timestamp,
				MerkleRoot:   identicalHeader.MerkleRoot,
				ReservedRoot: identicalHeader.ReservedRoot,
				Signature:    identicalHeader.Signature,
			},
			func(t *testing.T, err error) {
				require.Equal(t, ErrUpdateQueueIdenticalTimestamp, err)
			},
		},
		{
			"throttled",
			signUpdate(t, &wire.Update{
				Name:         throttledHeader.Name,
				Timestamp:    throttledHeader.Timestamp.Add(10 * time.Second),
				MerkleRoot:   throttledHeader.MerkleRoot,
				ReservedRoot: identicalHeader.ReservedRoot,
			}),
			func(t *testing.T, err error) {
				require.Equal(t, ErrUpdateQueueThrottled, err)
			},
		},
		{
			"stale",
			signUpdate(t, &wire.Update{
				Name:         staleHeader.Name,
				Timestamp:    staleHeader.Timestamp.Add(-10 * time.Second),
				MerkleRoot:   throttledHeader.MerkleRoot,
				ReservedRoot: identicalHeader.ReservedRoot,
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
		Name:       "somename",
		Timestamp:  time.Unix(100, 0),
		ReceivedAt: time.Unix(1, 0),
	})

	_, pub := testcrypto.FixedKey(t)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.SetInitialImportCompleteTx(tx); err != nil {
			return err
		}
		if err := store.SetNameInfoTx(tx, header.Name, pub, 10); err != nil {
			return err
		}
		if err := store.SetHeaderTx(tx, header, blob.ZeroMerkleBase); err != nil {
			return err
		}
		return nil
	}))

	queue := NewUpdateQueue(p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t)), db)
	require.NoError(t, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:      header.Name,
		Timestamp: header.Timestamp.Add(1 * time.Second),
	})))
	require.Equal(t, ErrUpdateQueueStaleTimestamp, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:      header.Name,
		Timestamp: header.Timestamp.Add(-10 * time.Second),
	})))
	require.Equal(t, ErrUpdateQueueSpltBrain, queue.Enqueue(crypto.Rand32(), signUpdate(t, &wire.Update{
		Name:       header.Name,
		Timestamp:  header.Timestamp.Add(1 * time.Second),
		MerkleRoot: crypto.Rand32(),
	})))
}

func TestUpdateQueue_EnqueueDequeue(t *testing.T) {
	db, done := setupDB(t)
	defer done()

	header := signHeader(t, &store.Header{
		Name:       "somename",
		Timestamp:  time.Unix(100, 0),
		ReceivedAt: time.Unix(1, 0),
	})

	_, pub := testcrypto.FixedKey(t)
	require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
		if err := store.SetInitialImportCompleteTx(tx); err != nil {
			return err
		}
		if err := store.SetNameInfoTx(tx, header.Name, pub, 10); err != nil {
			return err
		}
		if err := store.SetHeaderTx(tx, header, blob.ZeroMerkleBase); err != nil {
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
		Name:      header.Name,
		Timestamp: header.Timestamp.Add(time.Second),
	})
	for _, pid := range pids {
		require.NoError(t, queue.Enqueue(pid, update))
	}

	item := queue.Dequeue()
	for _, pid := range pids {
		require.True(t, item.PeerIDs.Has(pid))
	}
	require.Equal(t, update.Name, item.Name)
	require.Equal(t, update.Timestamp, item.Timestamp)
	require.Equal(t, update.MerkleRoot, item.MerkleRoot)
	require.Equal(t, update.ReservedRoot, item.ReservedRoot)
	require.Equal(t, update.Signature, item.Signature)
	require.True(t, pub.IsEqual(item.Pub))
	require.Equal(t, 10, item.Height)
}

func signHeader(t *testing.T, header *store.Header) *store.Header {
	sig, err := blob.SignSeal(testcrypto.FixedSigner(t), header.Name, header.Timestamp, header.MerkleRoot, header.ReservedRoot)
	require.NoError(t, err)
	header.Signature = sig
	return header
}

func signUpdate(t *testing.T, update *wire.Update) *wire.Update {
	sig, err := blob.SignSeal(testcrypto.FixedSigner(t), update.Name, update.Timestamp, update.MerkleRoot, update.ReservedRoot)
	require.NoError(t, err)
	update.Signature = sig
	return update
}
