package p2p

import (
	"context"
	"fmt"
	"fnd/crypto"
	"fnd/log"
	"fnd/service"
	"fnd/store"
	"fnd/util"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/sync/semaphore"
	"net"
	"sync"
	"time"
)

const (
	StandardPort = 9097

	MainnetMagic    = 0xcafecafe
	ProtocolVersion = 1

	MaxPendingInbound  = 12
	MaxPendingOutbound = 5

	DayBan  = 24 * time.Hour
	YearBan = 365 * DayBan
)

type PeerDialer interface {
	DialPeer(id crypto.Hash, ip string, verify bool) error
}

type PeerMeta struct {
	ID        crypto.Hash
	IP        string
	Port      int
	SentBytes int64
	RecvBytes int64
}

var (
	ErrMaxOutbound       = errors.New("reached maximum outbound peers")
	ErrMaxInbound        = errors.New("reached maximum inbound peers")
	ErrAlreadyConnecting = errors.New("already connecting to this peer")
	ErrPeerIDMismatch    = errors.New("peer IDs do not match after handshake")
	ErrSelfDial          = errors.New("self-dial after handshake")
	ErrAlreadyConnected  = errors.New("already connected to this peer")
	ErrPeerBanned        = errors.New("peer is banned")
	ErrInboundBusy       = errors.New("all inbound connections busy")
	ErrOutboundBusy      = errors.New("all outbound connections busy")
)

type PeerManager interface {
	service.Service
	PeerDialer
	AcceptPeer(conn *net.TCPConn) error
}

type peerManager struct {
	mux             *PeerMuxer
	db              *leveldb.DB
	maxInbound      int
	maxOutbound     int
	obs             *util.Observable
	signer          crypto.Signer
	listenHost      string
	magic           uint32
	protocolVersion uint32
	peerID          crypto.Hash
	pendingInbound  map[string]bool
	pendingOutbound map[string]bool
	doneCh          chan struct{}

	inSem  *semaphore.Weighted
	outSem *semaphore.Weighted
	inMu   sync.Mutex
	outMu  sync.Mutex
	lgr    log.Logger
}

type PeerManagerOpts struct {
	Mux         *PeerMuxer
	DB          *leveldb.DB
	SeedPeers   []SeedPeer
	Signer      crypto.Signer
	ListenHost  string
	MaxInbound  int
	MaxOutbound int
}

func NewPeerManager(opts *PeerManagerOpts) PeerManager {
	return &peerManager{
		maxInbound:      opts.MaxInbound,
		maxOutbound:     opts.MaxOutbound,
		obs:             util.NewObservable(),
		mux:             opts.Mux,
		db:              opts.DB,
		signer:          opts.Signer,
		listenHost:      opts.ListenHost,
		magic:           MainnetMagic,
		protocolVersion: ProtocolVersion,
		peerID:          crypto.HashPub(opts.Signer.Pub()),
		pendingInbound:  make(map[string]bool),
		pendingOutbound: make(map[string]bool),
		doneCh:          make(chan struct{}),
		inSem:           semaphore.NewWeighted(MaxPendingInbound),
		outSem:          semaphore.NewWeighted(MaxPendingOutbound),
		lgr:             log.WithModule("peer-manager"),
	}
}

func (p *peerManager) Start() error {
	p.lgr.Info("starting peer manager", "peer_id", p.peerID)

	go func() {
		refillTick := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-refillTick.C:
				p.refillPeers()
			case <-p.doneCh:
				return
			}
		}
	}()
	go p.refillPeers()
	return nil
}

func (p *peerManager) Stop() error {
	close(p.doneCh)
	return nil
}

func (p *peerManager) AcceptPeer(conn *net.TCPConn) error {
	if !p.inSem.TryAcquire(1) {
		if err := conn.Close(); err != nil {
			p.lgr.Error("error closing peer connection", "remote_addr", conn.RemoteAddr(), "err", err)
		}
		return ErrInboundBusy
	}
	defer p.inSem.Release(1)

	tcpAddr := conn.RemoteAddr().(*net.TCPAddr)
	if err := p.gateInboundPeer(tcpAddr); err != nil {
		if err := conn.Close(); err != nil {
			p.lgr.Error("error closing peer connection", "remote_addr", conn.RemoteAddr(), "err", err)
		}
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	peer := NewPeer(Inbound, conn)
	theirPeerID, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
		Magic:           p.magic,
		ProtocolVersion: p.protocolVersion,
		Peer:            peer,
		Signer:          p.signer,
	})
	if err != nil {
		if err := peer.Close(); err != nil {
			p.lgr.Error("error closing peer connection", "remote_addr", conn.RemoteAddr(), "err", err)
		}
		p.cleanupInboundPeer(tcpAddr.String())
		return err
	}
	return p.completeConnection(theirPeerID, peer, false)
}

func (p *peerManager) gateInboundPeer(addr *net.TCPAddr) error {
	p.inMu.Lock()
	defer p.inMu.Unlock()
	if addr.IP.String() == p.listenHost {
		return ErrSelfDial
	}
	if p.pendingInbound[addr.String()] {
		return ErrAlreadyConnecting
	}
	in, _ := p.mux.PeerCount()
	if in >= p.maxInbound {
		return ErrMaxInbound
	}
	isBanned, _, err := store.IsBanned(p.db, addr.IP.String())
	if err != nil {
		p.lgr.Error("error checking inbound ban state", "err", err)
	} else if isBanned {
		return ErrPeerBanned
	}
	p.pendingInbound[addr.String()] = true
	return nil
}

func (p *peerManager) cleanupInboundPeer(addr string) {
	p.inMu.Lock()
	defer p.inMu.Unlock()
	delete(p.pendingInbound, addr)
}

func (p *peerManager) DialPeer(peerID crypto.Hash, ip string, verifyPeerID bool) error {
	if !p.outSem.TryAcquire(1) {
		return ErrOutboundBusy
	}
	defer p.outSem.Release(1)
	if err := p.gateOutboundPeer(peerID, ip); err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, StandardPort), 2*time.Second)
	if err != nil {
		p.banOutboundPeer(ip)
		p.cleanupOutboundPeer(ip)
		return errors.Wrap(err, "dial failed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	peer := NewPeer(Outbound, conn)
	theirPeerID, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
		Magic:           p.magic,
		ProtocolVersion: p.protocolVersion,
		Peer:            peer,
		Signer:          p.signer,
	})
	if err != nil {
		_ = peer.Close()
		p.banOutboundPeer(ip)
		p.cleanupOutboundPeer(ip)
		return err
	}
	if verifyPeerID && peerID != theirPeerID {
		_ = peer.Close()
		p.banOutboundPeer(ip)
		p.cleanupOutboundPeer(ip)
		return ErrPeerIDMismatch
	}
	return p.completeConnection(theirPeerID, peer, verifyPeerID)
}

func (p *peerManager) gateOutboundPeer(peerID crypto.Hash, ip string) error {
	p.outMu.Lock()
	defer p.outMu.Unlock()
	if ip == p.listenHost {
		return ErrSelfDial
	}
	if p.pendingOutbound[ip] {
		return ErrAlreadyConnecting
	}
	_, out := p.mux.PeerCount()
	if out >= p.maxOutbound {
		return ErrMaxOutbound
	}
	if p.mux.HasPeerID(peerID) {
		return ErrAlreadyConnected
	}
	if p.mux.HasOutboundPeerIP(ip) {
		return ErrAlreadyConnected
	}
	_, isBanned, err := store.IsBanned(p.db, ip)
	if err != nil {
		p.lgr.Error("error checking outbound ban state", "err", err)
	} else if isBanned {
		return ErrPeerBanned
	}
	p.pendingOutbound[ip] = true
	return nil
}

func (p *peerManager) cleanupOutboundPeer(ip string) {
	p.outMu.Lock()
	defer p.outMu.Unlock()
	delete(p.pendingOutbound, ip)
}

func (p *peerManager) completeConnection(peerID crypto.Hash, peer Peer, verify bool) error {
	rIP := peer.RemoteIP()
	// these need to be deleted first
	if peer.Direction() == Inbound {
		p.cleanupInboundPeer(peer.RemoteAddr())
	} else {
		p.cleanupOutboundPeer(rIP)
	}

	if p.peerID == peerID {
		if err := peer.Close(); err != nil {
			p.lgr.Error("failed to close peer", "err", err)
		}
		// if we dialed ourselves, only one side of the connection
		// needs to ban anything.
		if peer.Direction() == Outbound {
			if err := store.BanOutboundPeer(p.db, rIP, YearBan); err != nil {
				p.lgr.Error("error banning peer after self-dial", "err", err)
			}
		}
		return ErrSelfDial
	}

	if err := p.mux.AddPeer(peerID, peer); err != nil {
		if err := peer.Close(); err != nil {
			p.lgr.Error("error closing peer", "peer_id", peerID, "err", err)
		}
		if errors.Is(err, ErrAlreadyConnected) {
			return err
		}
		return errors.Wrap(err, "error completing peer connection")
	}
	if err := store.SetPeer(p.db, peerID, rIP, verify); err != nil {
		p.lgr.Error("error saving peer", "err", err)
	}
	p.lgr.Info("peer added", "peer_id", peerID, "direction", peer.Direction())
	return nil
}

func (p *peerManager) refillPeers() {
	_, outCount := p.mux.PeerCount()
	if outCount >= p.maxOutbound {
		return
	}

	p.lgr.Info("refilling peers", "have", outCount, "want", p.maxOutbound)
	peerStream, err := store.StreamPeers(p.db, false)
	if err != nil {
		p.lgr.Error("error opening peer stream", "err", err)
		return
	}
	// can't stream and modify the DB at the same time
	var candidatePeers []*store.Peer
	for len(candidatePeers) < 256 {
		peer, err := peerStream.Next()
		if err != nil {
			p.lgr.Error("error streaming stored peer", "err", err)
			break
		}
		if peer == nil {
			break
		}
		if p.mux.HasPeerID(peer.ID) {
			continue
		}
		candidatePeers = append(candidatePeers, peer)
	}
	peerStream.Close()

	for _, peer := range candidatePeers {
		_, outCount := p.mux.PeerCount()
		if outCount >= p.maxOutbound {
			break
		}

		if err := p.DialPeer(peer.ID, peer.IP, peer.Verify); err != nil {
			p.lgr.Error("failed to dial peer during refill", "err", err)
			continue
		}
	}
}

func (p *peerManager) banOutboundPeer(ip string) {
	// TODO: avoid banning seed peers
	if err := store.BanOutboundPeer(p.db, ip, time.Hour); err != nil {
		p.lgr.Error("error storing outbound peer ban state", "err", err)
	}
}
