package protocol

import (
	"github.com/ddrp-org/ddrp/crypto"
	"sync"
)

type PeerSet struct {
	peers map[crypto.Hash]bool
	ids   []crypto.Hash
	mtx   sync.RWMutex
}

func NewPeerSet(initialPeers []crypto.Hash) *PeerSet {
	res := &PeerSet{
		peers: make(map[crypto.Hash]bool),
	}
	for _, peer := range initialPeers {
		if res.peers[peer] {
			continue
		}
		res.peers[peer] = true
		res.ids = append(res.ids, peer)
	}
	return res
}

func (ps *PeerSet) Add(peerID crypto.Hash) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	if ps.peers[peerID] {
		return
	}
	ps.peers[peerID] = true
	ps.ids = append(ps.ids, peerID)
}

func (ps *PeerSet) Has(peerID crypto.Hash) bool {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return ps.peers[peerID]
}

func (ps *PeerSet) Len() int {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	return len(ps.peers)
}

func (ps *PeerSet) Iterator() func() (crypto.Hash, bool) {
	var i int
	return func() (crypto.Hash, bool) {
		ps.mtx.RLock()
		defer ps.mtx.RUnlock()
		if i >= len(ps.ids) {
			return crypto.ZeroHash, false
		}
		j := i
		i++
		return ps.ids[j], true
	}
}
