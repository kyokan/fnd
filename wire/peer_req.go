package wire

import (
	"ddrp/crypto"
	"io"
)

type PeerReq struct {
	HashCacher
}

var _ Message = (*PeerReq)(nil)

func (p *PeerReq) MsgType() MessageType {
	return MessageTypePeerReq
}

func (p *PeerReq) Equals(other Message) bool {
	_, ok := other.(*PeerReq)
	return ok
}

func (p *PeerReq) Encode(w io.Writer) error {
	return nil
}

func (p *PeerReq) Decode(r io.Reader) error {
	return nil
}

func (p *PeerReq) Hash() (crypto.Hash, error) {
	return p.HashCacher.Hash(p)
}
