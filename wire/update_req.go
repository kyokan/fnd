package wire

import (
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
	"time"
)

type UpdateReq struct {
	HashCacher

	Name      string
	Timestamp time.Time
}

var _ Message = (*UpdateReq)(nil)

func (n *UpdateReq) MsgType() MessageType {
	return MessageTypeUpdateReq
}

func (n *UpdateReq) Equals(other Message) bool {
	cast, ok := other.(*UpdateReq)
	if !ok {
		return false
	}

	return n.Name == cast.Name &&
		n.Timestamp.Unix() == cast.Timestamp.Unix()
}

func (n *UpdateReq) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		n.Name,
		n.Timestamp,
	)
}

func (n *UpdateReq) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&n.Name,
		&n.Timestamp,
	)
}

func (n *UpdateReq) Hash() (crypto.Hash, error) {
	return n.HashCacher.Hash(n)
}
