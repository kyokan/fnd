package wire

import (
	"bytes"
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
	"net"
)

type Peer struct {
	IP net.IP
	ID crypto.Hash
}

func (p *Peer) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		IPEncoder(p.IP),
		p.ID,
	)
}

func (p *Peer) Decode(r io.Reader) error {
	var ip IPEncoder
	err := dwire.DecodeFields(
		r,
		&ip,
		&p.ID,
	)
	if err != nil {
		return err
	}
	p.IP = net.IP(ip)
	return nil
}

type PeerRes struct {
	HashCacher

	Peers []*Peer
}

func (p *PeerRes) MsgType() MessageType {
	return MessageTypePeerRes
}

func (p *PeerRes) Equals(other Message) bool {
	cast, ok := other.(*PeerRes)
	if !ok {
		return false
	}

	for i := 0; i < len(p.Peers); i++ {
		peerA := p.Peers[i]
		peerB := cast.Peers[i]
		isEqual := peerA.IP.Equal(peerB.IP) &&
			peerA.ID == peerB.ID
		if !isEqual {
			return false
		}
	}

	return true
}

func (p *PeerRes) Encode(w io.Writer) error {
	var buf bytes.Buffer
	for _, peer := range p.Peers {
		if err := peer.Encode(&buf); err != nil {
			return err
		}
	}

	return dwire.EncodeFields(
		w,
		p.Peers,
	)
}

func (p *PeerRes) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&p.Peers,
	)
}

func (p *PeerRes) Hash() (crypto.Hash, error) {
	return p.HashCacher.Hash(p)
}
