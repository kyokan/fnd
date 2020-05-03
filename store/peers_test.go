package store

import (
	"ddrp/crypto"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
)

func TestPeers(t *testing.T) {
	db, done := setupLevelDB(t)

	idsIPs := map[crypto.Hash]string{
		crypto.Rand32(): "127.0.0.1",
		crypto.Rand32(): "127.0.0.2",
		crypto.Rand32(): "127.0.0.3",
	}
	ids := make([]crypto.Hash, 0)
	for id := range idsIPs {
		ids = append(ids, id)
	}

	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		for id, ip := range idsIPs {
			if err := SetPeerTx(tx, id, ip, true); err != nil {
				return err
			}
		}
		return nil
	}))

	for id := range idsIPs {
		inBan, outBan, err := IsBanned(db, idsIPs[id])
		require.NoError(t, err)
		require.False(t, inBan)
		require.False(t, outBan)
	}

	streamedPeers := getAllPeers(t, db, true)
	require.Equal(t, len(ids), len(streamedPeers))

	dur := 10 * time.Minute
	require.NoError(t, WithTx(db, func(tx *leveldb.Transaction) error {
		if err := BanInboundPeerTx(tx, idsIPs[ids[0]], dur); err != nil {
			return err
		}
		if err := BanOutboundPeerTx(tx, idsIPs[ids[1]], dur); err != nil {
			return err
		}
		return nil
	}))

	inBan, outBan, err := IsBanned(db, idsIPs[ids[0]])
	require.NoError(t, err)
	require.True(t, inBan)
	require.False(t, outBan)

	inBan, outBan, err = IsBanned(db, idsIPs[ids[1]])
	require.NoError(t, err)
	require.False(t, inBan)
	require.True(t, outBan)

	streamedPeers = getAllPeers(t, db, true)
	require.Equal(t, len(ids), len(streamedPeers))

	streamedPeers = getAllPeers(t, db, false)
	require.Equal(t, 1, len(streamedPeers))

	done()
}

func getAllPeers(t *testing.T, db *leveldb.DB, includeBanned bool) []*Peer {
	var out []*Peer
	stream, err := StreamPeers(db, includeBanned)
	require.NoError(t, err)
	for {
		peer, err := stream.Next()
		require.NoError(t, err)
		if peer == nil {
			break
		}
		out = append(out, peer)
	}
	require.NoError(t, stream.Close())
	return out
}
