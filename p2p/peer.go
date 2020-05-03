package p2p

import (
	"bufio"
	"context"
	"ddrp/log"
	"ddrp/wire"
	"errors"
	"golang.org/x/time/rate"
	"io"
	"math"
	"net"
	"sync"
	"time"
)

var (
	ErrPeerSendBufferFull = errors.New("peer send buffer full")
	ErrPeerRecvBufferFull = errors.New("peer receive buffer full")
)

type PeerDirection int

func (p PeerDirection) String() string {
	switch p {
	case Inbound:
		return "Inbound"
	case Outbound:
		return "Outbound"
	default:
		panic("invalid peer direction")
	}
}

var (
	ErrPeerClosed = errors.New("peer closed")
	ErrPeerHangup = errors.New("remote hung up")
)

const (
	DefaultPeerRecvRateLimit      = 64
	DefaultPeerRecvRateLimitBurst = 192
	MaxPeerPacketSize             = 5 * 1024 * 1024
	IdleTimeout                   = 10 * time.Second
	ValidMessageDeadline          = 2 * time.Hour

	Inbound  PeerDirection = 0
	Outbound PeerDirection = 1
)

type Peer interface {
	Direction() PeerDirection
	LocalAddr() string
	RemoteIP() string
	RemoteAddr() string
	RemotePort() int
	SendCtx(ctx context.Context, envelope *wire.Envelope) error
	Send(envelope *wire.Envelope) error
	ReceiveCtx(ctx context.Context) (*wire.Envelope, error)
	Receive() (*wire.Envelope, error)
	CloseChan() <-chan struct{}
	Close() error
	BandwidthUsage() (uint64, uint64)
	CloseReason() error
}

type PeerImpl struct {
	direction  PeerDirection
	conn       net.Conn
	connW      *CountingWriter
	connR      *CountingReader
	listenPort int
	lgr        log.Logger

	sendCh        chan *sendReq
	recvCh        chan *recvReq
	sendDoneCh    chan struct{}
	recvDoneCh    chan struct{}
	closeCh       chan struct{}
	closeMu       sync.Mutex
	closeReason   error
	closeReasonMu sync.Mutex
}

type sendReq struct {
	envelope *wire.Envelope
	errCh    chan error
}

type recvReq struct {
	envelopeCh chan *wire.Envelope
	errCh      chan error
}

func NewPeer(direction PeerDirection, conn net.Conn) Peer {
	p := &PeerImpl{
		direction:  direction,
		conn:       conn,
		connW:      NewCountingWriter(conn),
		connR:      NewCountingReader(bufio.NewReader(conn)),
		sendCh:     make(chan *sendReq, 128),
		recvCh:     make(chan *recvReq, 128),
		sendDoneCh: make(chan struct{}, 1),
		recvDoneCh: make(chan struct{}, 1),
		closeCh:    make(chan struct{}),
		lgr:        log.WithModule("peer").Sub("remote_addr", conn.RemoteAddr()),
	}
	go p.send()
	go p.recv()
	return p
}

func (p *PeerImpl) SendCtx(ctx context.Context, envelope *wire.Envelope) error {
	select {
	case <-p.closeCh:
		return p.CloseReason()
	default:
	}

	errCh := make(chan error, 1)
	req := &sendReq{
		envelope: envelope,
		errCh:    errCh,
	}
	if err := p.bufferSendCtx(ctx, req); err != nil {
		return err
	}

	select {
	case err := <-errCh:
		return err
	case <-p.closeCh:
		return p.CloseReason()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *PeerImpl) Send(envelope *wire.Envelope) error {
	return p.SendCtx(context.Background(), envelope)
}

func (p *PeerImpl) ReceiveCtx(ctx context.Context) (*wire.Envelope, error) {
	select {
	case <-p.closeCh:
		return nil, p.CloseReason()
	default:
	}

	envelopeCh := make(chan *wire.Envelope, 1)
	errCh := make(chan error, 1)
	req := &recvReq{
		envelopeCh: envelopeCh,
		errCh:      errCh,
	}

	select {
	case p.recvCh <- req:
	default:
		return nil, ErrPeerRecvBufferFull
	}

	select {
	case envelope := <-envelopeCh:
		return envelope, nil
	case err := <-errCh:
		return nil, err
	case <-p.closeCh:
		return nil, p.CloseReason()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *PeerImpl) Receive() (*wire.Envelope, error) {
	return p.ReceiveCtx(context.Background())
}

func (p *PeerImpl) CloseChan() <-chan struct{} {
	return p.closeCh
}

func (p *PeerImpl) Close() error {
	p.closeMu.Lock()
	defer p.closeMu.Unlock()

	select {
	case <-p.closeCh:
		return nil
	default:
	}

	p.closeReasonMu.Lock()
	if p.closeReason == nil {
		p.closeReason = ErrPeerClosed
	}
	p.closeReasonMu.Unlock()
	_ = p.conn.Close()
	close(p.closeCh)
	<-p.recvDoneCh
	<-p.sendDoneCh
	return nil
}

func (p *PeerImpl) Direction() PeerDirection {
	return p.direction
}

func (p *PeerImpl) LocalAddr() string {
	return p.conn.LocalAddr().String()
}

func (p *PeerImpl) RemoteIP() string {
	return p.conn.RemoteAddr().(*net.TCPAddr).IP.String()
}

func (p *PeerImpl) RemoteAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *PeerImpl) RemotePort() int {
	return p.conn.RemoteAddr().(*net.TCPAddr).Port
}

func (p *PeerImpl) BandwidthUsage() (uint64, uint64) {
	return p.connW.Count(), p.connR.Count()
}

func (p *PeerImpl) CloseReason() error {
	p.closeReasonMu.Lock()
	defer p.closeReasonMu.Unlock()
	return p.closeReason
}

func (p *PeerImpl) send() {
	defer func() {
		p.sendDoneCh <- struct{}{}
		_ = p.Close()
	}()

	for {
		select {
		case sendReq := <-p.sendCh:
			envelope := sendReq.envelope
			p.updateDeadline()
			if err := envelope.Encode(p.connW); err != nil {
				sendReq.errCh <- p.setCloseReason(err)
				return
			}
			sendReq.errCh <- nil
			p.lgr.Trace("sent message", "message_type", envelope.MessageType)
		case <-p.closeCh:
			return
		}
	}
}

func (p *PeerImpl) recv() {
	defer func() {
		p.recvDoneCh <- struct{}{}
		_ = p.Close()
	}()

	lim := rate.NewLimiter(DefaultPeerRecvRateLimit, DefaultPeerRecvRateLimitBurst)
	for {
		select {
		case recvReq := <-p.recvCh:
			rv := lim.Reserve()
			if !rv.OK() {
				time.Sleep(rv.DelayFrom(time.Now()))
			}
			p.updateDeadline()
			envelope := new(wire.Envelope)
			err := envelope.Decode(io.LimitReader(p.connR, MaxPeerPacketSize))
			if err != nil {
				recvReq.errCh <- p.setCloseReason(err)
				return
			}

			p.lgr.Trace("received message", "message_type", envelope.MessageType)
			recvReq.envelopeCh <- envelope
		case <-p.closeCh:
			return
		}
	}
}

func (p *PeerImpl) setCloseReason(err error) error {
	p.closeReasonMu.Lock()
	defer p.closeReasonMu.Unlock()
	if p.closeReason != nil {
		return p.closeReason
	}
	if err == nil {
		return nil
	}
	if err == io.EOF {
		p.closeReason = ErrPeerHangup
	} else {
		p.closeReason = err
	}
	return p.closeReason
}

func (p *PeerImpl) updateDeadline() {
	_ = p.conn.SetDeadline(time.Now().Add(time.Minute))
}

func (p *PeerImpl) bufferSendCtx(ctx context.Context, req *sendReq) error {
	var retries int
	for {
		if retries > 10 {
			return ErrPeerSendBufferFull
		}

		timer := time.NewTimer(time.Duration(int(math.Pow(2, float64(retries)))) * time.Millisecond)
		select {
		case p.sendCh <- req:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			retries++
			continue
		}
	}
}
