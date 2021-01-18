package p2p

import (
	"fmt"
	"fnd/crypto"
	"fnd/log"
	"fnd/util"
	"fnd/wire"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

const (
	DefaultPeerMuxerGossipTimeoutMS = 5 * 60 * 1000
)

type PeerMessageHandler func(peerID crypto.Hash, envelope *wire.Envelope)
type PeerStateHandler func(peerID crypto.Hash)

type PeerMuxer struct {
	GossipTimeoutMS   int
	outboundPeersByIP map[string]Peer
	inboundPeersByIP  map[string][]Peer
	peers             map[crypto.Hash]Peer
	obs               *util.Observable
	mu                sync.RWMutex
	gossipFilter      *util.Cache
	inboundCount      int
	outboundCount     int
	magic             uint32
	signer            crypto.Signer
	bytesTx           uint64
	bytesRx           uint64
	lgr               log.Logger
}

func NewPeerMuxer(magic uint32, signer crypto.Signer) *PeerMuxer {
	return &PeerMuxer{
		GossipTimeoutMS:   DefaultPeerMuxerGossipTimeoutMS,
		outboundPeersByIP: make(map[string]Peer),
		inboundPeersByIP:  make(map[string][]Peer),
		peers:             make(map[crypto.Hash]Peer),
		obs:               util.NewObservable(),
		gossipFilter:      util.NewCache(),
		magic:             magic,
		signer:            signer,
		lgr:               log.WithModule("peer-muxer"),
	}
}

func (p *PeerMuxer) BandwidthUsage() (uint64, uint64) {
	return atomic.LoadUint64(&p.bytesTx), atomic.LoadUint64(&p.bytesRx)
}

func (p *PeerMuxer) PeerCount() (int, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.inboundCount, p.outboundCount
}

func (p *PeerMuxer) HasPeerID(id crypto.Hash) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.peers[id]
	return ok
}

func (p *PeerMuxer) HasOutboundPeerIP(ip string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.outboundPeersByIP[ip]
	return ok
}

func (p *PeerMuxer) Peers() map[crypto.Hash]Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	peers := make(map[crypto.Hash]Peer)
	for peerID, peer := range p.peers {
		peers[peerID] = peer
	}
	return peers
}

func (p *PeerMuxer) PeersByIP(ip string) []Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var peers []Peer
	for _, peer := range p.inboundPeersByIP[ip] {
		peers = append(peers, peer)
	}
	outbound := p.outboundPeersByIP[ip]
	if outbound != nil {
		peers = append(peers, outbound)
	}
	return peers
}

func (p *PeerMuxer) PeerIDs() []crypto.Hash {
	var out []crypto.Hash
	for id := range p.peers {
		out = append(out, id)
	}
	return out
}

func (p *PeerMuxer) PeerByID(id crypto.Hash) (Peer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	peer := p.peers[id]
	if peer == nil {
		return nil, errors.New("peer not found")
	}
	return peer, nil
}

func (p *PeerMuxer) PeerByIP(ip string) (Peer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	peer := p.outboundPeersByIP[ip]
	if peer == nil {
		return nil, errors.New("peer not found")
	}
	return peer, nil
}

func (p *PeerMuxer) GossipPeerIDs(message wire.Message) []crypto.Hash {
	peers := p.PeerIDs()
	var out []crypto.Hash
	for _, peerID := range peers {
		key := p.gossipKey(peerID, message)
		if p.gossipFilter.Has(key) {
			continue
		}

		out = append(out, peerID)
	}

	return out
}

func (p *PeerMuxer) AddPeer(id crypto.Hash, peer Peer) error {
	if err := p.handlePeerOpen(id, peer); err != nil {
		return errors.Wrap(err, "error adding peer")
	}
	go func() {
		for {
			_, startRx := peer.BandwidthUsage()
			envelope, err := peer.Receive()
			if err != nil {
				p.handlePeerClose(id)
				return
			}
			if err := ValidateEnvelope(p.magic, id, envelope); err != nil {
				p.lgr.Error("envelope failed validation, closing peer", "err", err)
				p.handlePeerClose(id)
				return
			}
			p.handlePeerMessage(id, envelope)
			_, endRx := peer.BandwidthUsage()
			atomic.AddUint64(&p.bytesRx, endRx-startRx)
		}
	}()
	return nil
}

func (p *PeerMuxer) ClosePeer(id crypto.Hash) error {
	p.handlePeerClose(id)
	return nil
}

func (p *PeerMuxer) AddMessageHandler(handler PeerMessageHandler) util.Unsubscriber {
	return p.obs.On("message", handler)
}

func (p *PeerMuxer) AddPeerOpenHandler(handler PeerStateHandler) util.Unsubscriber {
	return p.obs.On("open", handler)
}

func (p *PeerMuxer) Send(id crypto.Hash, message wire.Message) error {
	p.mu.RLock()
	peer, ok := p.peers[id]
	if !ok {
		p.mu.RUnlock()
		return errors.New("peer not found")
	}
	p.mu.RUnlock()

	envelope, err := wire.NewEnvelope(p.magic, message, p.signer)
	if err != nil {
		return errors.Wrap(err, "error creating envelope")
	}
	startTx, _ := peer.BandwidthUsage()
	if err := peer.Send(envelope); err != nil {
		return errors.Wrap(err, "error sending message to peer")
	}
	endTx, _ := peer.BandwidthUsage()
	atomic.AddUint64(&p.bytesTx, endTx-startTx)
	return nil
}

func (p *PeerMuxer) handlePeerMessage(id crypto.Hash, envelope *wire.Envelope) {
	p.gossipFilter.Set(p.gossipKey(id, envelope.Message), true, int64(p.GossipTimeoutMS))
	p.obs.Emit("message", id, envelope)
}

func (p *PeerMuxer) handlePeerClose(id crypto.Hash) {
	p.mu.Lock()
	peer := p.peers[id]
	if peer == nil {
		p.mu.Unlock()
		return
	}

	if peer.Direction() == Inbound {
		p.inboundCount--
		p.removeInboundPeerByIP(peer)
	} else {
		p.outboundCount--
		delete(p.outboundPeersByIP, peer.RemoteIP())
	}
	delete(p.peers, id)
	p.mu.Unlock()

	peer.Close()
	p.obs.Emit("close", id)
	p.lgr.Info("peer closed", "peer_id", id, "reason", peer.CloseReason())
}

func (p *PeerMuxer) handlePeerOpen(id crypto.Hash, peer Peer) error {
	p.mu.Lock()
	_, ok := p.peers[id]
	if ok {
		p.mu.Unlock()
		return ErrAlreadyConnected
	}
	if peer.Direction() == Inbound {
		p.inboundCount++
		p.inboundPeersByIP[peer.RemoteIP()] = append(p.inboundPeersByIP[peer.RemoteIP()], peer)
	} else {
		p.outboundCount++
		p.outboundPeersByIP[peer.RemoteIP()] = peer
	}
	p.peers[id] = peer
	p.mu.Unlock()

	p.obs.Emit("open", id)
	return nil
}

func (p *PeerMuxer) gossipKey(peerID crypto.Hash, message wire.Message) string {
	hash, _ := message.Hash()
	return fmt.Sprintf("%s:%s", peerID, hash)
}

func (p *PeerMuxer) removeInboundPeerByIP(peer Peer) {
	for i, storedPeer := range p.inboundPeersByIP[peer.RemoteIP()] {
		if storedPeer == peer {
			p.inboundPeersByIP[peer.RemoteIP()] = append(p.inboundPeersByIP[peer.RemoteIP()][:i], p.inboundPeersByIP[peer.RemoteIP()][i+1:]...)
		}
	}
}

func PeerMessageHandlerForType(msgType wire.MessageType, hdlr func(id crypto.Hash, envelope *wire.Envelope)) PeerMessageHandler {
	return func(id crypto.Hash, envelope *wire.Envelope) {
		if envelope.MessageType != msgType {
			return
		}
		hdlr(id, envelope)
	}
}

func BroadcastRandom(mux *PeerMuxer, size int, message wire.Message) ([]crypto.Hash, []error) {
	peerIDs := mux.PeerIDs()
	indices := util.SampleIndices(len(peerIDs), size)
	var recips []crypto.Hash
	var errs []error
	for _, i := range indices {
		peerID := peerIDs[i]
		recips = append(recips, peerID)
		errs = append(errs, mux.Send(peerID, message))
	}
	return recips, errs
}

func BroadcastAll(mux *PeerMuxer, message wire.Message) ([]crypto.Hash, []error) {
	recips := mux.PeerIDs()
	var errs []error
	for _, recip := range recips {
		errs = append(errs, mux.Send(recip, message))
	}
	return recips, errs
}

func GossipAll(mux *PeerMuxer, message wire.Message) ([]crypto.Hash, []error) {
	recips := mux.GossipPeerIDs(message)
	var errs []error
	for _, recip := range recips {
		errs = append(errs, mux.Send(recip, message))
	}
	return recips, errs
}
