package wire

import (
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
)

type NilUpdate struct {
	HashCacher

	Name string

	hash crypto.Hash
}

var _ Message = (*NilUpdate)(nil)

func NewNilUpdate(name string) *NilUpdate {
	return &NilUpdate{
		Name: name,
	}
}

func (n *NilUpdate) MsgType() MessageType {
	return MessageTypeNilUpdate
}

func (n *NilUpdate) Equals(other Message) bool {
	cast, ok := other.(*NilUpdate)
	if !ok {
		return false
	}
	return n.Name == cast.Name
}

func (n *NilUpdate) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		n.Name,
	)
}

func (n *NilUpdate) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&n.Name,
	)
}

func (n *NilUpdate) Hash() (crypto.Hash, error) {
	return n.HashCacher.Hash(n)
}
