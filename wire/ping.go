package wire

import (
	"fnd/crypto"
	"io"
)

type Ping struct {
	HashCacher
}

var _ Message = (*Ping)(nil)

func NewPing() *Ping {
	return &Ping{}
}

func (p *Ping) MsgType() MessageType {
	return MessageTypePing
}

func (p *Ping) Equals(other Message) bool {
	_, ok := other.(*Ping)
	return ok
}

func (p *Ping) Encode(w io.Writer) error {
	return nil
}

func (p *Ping) Decode(r io.Reader) error {
	return nil
}

func (p *Ping) Hash() (crypto.Hash, error) {
	return p.HashCacher.Hash(p)
}
