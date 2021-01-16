package protocol

import (
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPeerSet(t *testing.T) {
	peerIDs := []crypto.Hash{
		crypto.Rand32(),
		crypto.Rand32(),
	}

	ps := NewPeerSet(peerIDs[:1])
	require.True(t, ps.Has(peerIDs[0]))
	require.False(t, ps.Has(peerIDs[1]))
	ps.Add(peerIDs[1])
	require.True(t, ps.Has(peerIDs[1]))

	iter := ps.Iterator()
	for {
		peerID, ok := iter()
		if !ok {
			break
		}
		require.Equal(t, peerIDs[0], peerID)
		peerIDs = peerIDs[1:]
	}
	require.Equal(t, 0, len(peerIDs))
}
