package wire

import (
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
)

type HelloAck struct {
	HashCacher

	Nonce [32]byte
}

var _ Message = (*HelloAck)(nil)

func (h *HelloAck) MsgType() MessageType {
	return MessageTypeHelloAck
}

func (h *HelloAck) Equals(other Message) bool {
	cast, ok := other.(*HelloAck)
	if !ok {
		return false
	}

	return h.Nonce == cast.Nonce
}

func (h *HelloAck) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		h.Nonce,
	)
}

func (h *HelloAck) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&h.Nonce,
	)
}

func (h *HelloAck) Hash() (crypto.Hash, error) {
	return h.HashCacher.Hash(h)
}
