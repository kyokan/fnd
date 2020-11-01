package protocol

import (
	"fnd/config"
	"fnd/crypto"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/util"
	"fnd/wire"
	"github.com/syndtr/goleveldb/leveldb"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	MaxSentPeerCount = math.MaxUint8
)

type PeerExchanger struct {
	SampleSize      int
	ResponseTimeout time.Duration
	RequestInterval time.Duration
	dialer          p2p.PeerDialer
	mux             *p2p.PeerMuxer
	db              *leveldb.DB
	cache           *util.Cache
	activeDials     map[crypto.Hash]bool
	mtx             sync.Mutex
	lgr             log.Logger
	doneCh          chan struct{}
	obs             *util.Observable
}

func NewPeerExchanger(dialer p2p.PeerDialer, mux *p2p.PeerMuxer, db *leveldb.DB) *PeerExchanger {
	return &PeerExchanger{
		SampleSize:      config.DefaultConfig.Tuning.PeerExchanger.SampleSize,
		ResponseTimeout: config.ConvertDuration(config.DefaultConfig.Tuning.PeerExchanger.ResponseTimeoutMS, time.Millisecond),
		RequestInterval: config.ConvertDuration(config.DefaultConfig.Tuning.PeerExchanger.RequestIntervalMS, time.Millisecond),
		dialer:          dialer,
		mux:             mux,
		db:              db,
		cache:           util.NewCache(),
		activeDials:     make(map[crypto.Hash]bool),
		doneCh:          make(chan struct{}),
		obs:             util.NewObservable(),
		lgr:             log.WithModule("peer-exchanger"),
	}
}

func (pe *PeerExchanger) Start() error {
	pe.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypePeerReq, pe.handlePeerReq))
	pe.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypePeerRes, pe.handlePeerRes))

	tick := time.NewTicker(pe.RequestInterval)

	for {
		select {
		case <-tick.C:
			recips, _ := p2p.BroadcastAll(pe.mux, &wire.PeerReq{})
			for _, recip := range recips {
				pe.cache.Set(recip.String(), true, int64(pe.ResponseTimeout/time.Millisecond))
			}
			pe.lgr.Debug("requested new peers", "recipient_count", len(recips))
		case <-pe.doneCh:
			return nil
		}
	}
}

func (pe *PeerExchanger) Stop() error {
	close(pe.doneCh)
	return nil
}

func (pe *PeerExchanger) handlePeerReq(peerID crypto.Hash, envelope *wire.Envelope) {
	peerStream, err := store.StreamPeers(pe.db, false)
	if err != nil {
		pe.lgr.Error("error opening peer stream", "err", err)
		return
	}
	defer peerStream.Close()

	var peers []*wire.Peer
	for {
		if len(peers) == MaxSentPeerCount {
			break
		}
		peer, err := peerStream.Next()
		if err != nil {
			pe.lgr.Error("error streaming stored peer", "err", err)
			return
		}
		if peer == nil {
			break
		}
		peers = append(peers, &wire.Peer{
			IP: net.ParseIP(peer.IP),
			ID: peer.ID,
		})
	}

	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	msg := &wire.PeerRes{
		Peers: peers,
	}
	if err := pe.mux.Send(peerID, msg); err != nil {
		pe.lgr.Error("error sending peers", "err", err)
		return
	}
	pe.lgr.Info("sent peers to requestor", "count", len(msg.Peers), "peer_id", peerID)
}

func (pe *PeerExchanger) handlePeerRes(peerID crypto.Hash, envelope *wire.Envelope) {
	peerIDStr := peerID.String()
	if !pe.cache.Has(peerIDStr) {
		pe.lgr.Warn("received unsolicited PeerRes message")
		return
	}

	pe.cache.Del(peerIDStr)
	msg := envelope.Message.(*wire.PeerRes)
	pe.lgr.Debug("received new peers", "source_peer_id", peerID, "count", len(msg.Peers))
	for _, peer := range msg.Peers {
		pe.dialPeer(peer)
	}
}

func (pe *PeerExchanger) dialPeer(peer *wire.Peer) {
	peerID := peer.ID
	ipStr := peer.IP.String()
	pe.lgr.Trace("dialing exchanged peer", "ip", ipStr, "peer_id", peerID)
	err := pe.dialer.DialPeer(peerID, ipStr, false)
	if err == p2p.ErrAlreadyConnecting {
		pe.lgr.Trace("already connecting to exchanged peer", "ip", ipStr)
		return
	}
	if err == p2p.ErrAlreadyConnected {
		pe.lgr.Trace("already connected to exchanged peer", "ip", ipStr)
		return
	}
	if err == p2p.ErrPeerBanned {
		pe.lgr.Trace("peer is banned", "ip", ipStr)
		return
	}
	if err == p2p.ErrMaxOutbound {
		pe.lgr.Trace("at max outbound peers")
		return
	}
	if err == p2p.ErrSelfDial {
		pe.lgr.Trace("self-dial")
		return
	}
	if err != nil {
		pe.lgr.Error("failed to connect to exchanged peer", "ip", ipStr, "err", err)
	}
}
