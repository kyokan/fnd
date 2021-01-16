package protocol

import (
	"context"
	"errors"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/wire"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultPingerCheckInterval = time.Second
	DefaultPingerPingInterval  = 5 * time.Second
	DefaultPingerTimeout       = 30 * time.Second
)

type Pinger struct {
	CheckInterval time.Duration
	PingInterval  time.Duration
	Timeout       time.Duration
	mux           *p2p.PeerMuxer
	quitCh        chan struct{}
	wg            sync.WaitGroup
	lgr           log.Logger
}

var (
	ErrPingTimeout    = errors.New("ping timeout")
	ErrPingPeerClosed = errors.New("peer closed")

	pingLogger = log.WithModule("pinger")
)

func NewPinger(mux *p2p.PeerMuxer) *Pinger {
	return &Pinger{
		CheckInterval: DefaultPingerCheckInterval,
		PingInterval:  DefaultPingerPingInterval,
		Timeout:       DefaultPingerTimeout,
		mux:           mux,
		quitCh:        make(chan struct{}),
		lgr:           log.WithModule("pinger"),
	}
}

func (p *Pinger) Start() error {
	p.mux.AddPeerOpenHandler(p.handlePeerOpen)
	return nil
}

func (p *Pinger) Stop() error {
	close(p.quitCh)
	p.wg.Wait()
	return nil
}

func (p *Pinger) handlePeerOpen(peerID crypto.Hash) {
	p.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-p.quitCh:
			cancel()
			return
		}
	}()
	go func() {
		_ = PingPeer(ctx, &PingConfig{
			CheckInterval: p.CheckInterval,
			PingInterval:  p.PingInterval,
			Timeout:       p.Timeout,
			PeerID:        peerID,
			Mux:           p.mux,
		})
		cancel()
		p.wg.Done()
	}()
}

type PingConfig struct {
	CheckInterval time.Duration
	PingInterval  time.Duration
	Timeout       time.Duration
	PeerID        crypto.Hash
	Mux           *p2p.PeerMuxer
}

func PingPeer(ctx context.Context, cfg *PingConfig) error {
	lastPing := time.Now().UnixNano()
	unsubPing := cfg.Mux.AddMessageHandler(func(peerID crypto.Hash, envelope *wire.Envelope) {
		if cfg.PeerID != peerID {
			return
		}
		atomic.StoreInt64(&lastPing, time.Now().UnixNano())
	})
	defer unsubPing()

	checkTick := time.NewTicker(cfg.CheckInterval)
	pingTick := time.NewTicker(cfg.PingInterval)
	for {
		select {
		case <-checkTick.C:
			if !cfg.Mux.HasPeerID(cfg.PeerID) {
				pingLogger.Debug("peer is closed, shutting down pinger")
				return ErrPingPeerClosed
			}
			if time.Duration(time.Now().UnixNano()-atomic.LoadInt64(&lastPing)) > cfg.Timeout {
				pingLogger.Info("no pings received within timeout, closing peer")
				if err := cfg.Mux.ClosePeer(cfg.PeerID); err != nil {
					pingLogger.Error("failed to close peer after timeout", "peer_id", cfg.PeerID, "err", err)
				}
				return ErrPingTimeout
			}
		case <-pingTick.C:
			if !cfg.Mux.HasPeerID(cfg.PeerID) {
				pingLogger.Debug("peer is closed, shutting down pinger")
				return ErrPingPeerClosed
			}
			if err := cfg.Mux.Send(cfg.PeerID, wire.NewPing()); err != nil {
				pingLogger.Error("error sending ping", "peer_id", cfg.PeerID, "err", err)
				continue
			}
			pingLogger.Trace("sent ping", "peer_id", cfg.PeerID)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
