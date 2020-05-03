package wire

import (
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
)

type TreeBaseReq struct {
	HashCacher

	Name string
}

var _ Message = (*TreeBaseReq)(nil)

func (d *TreeBaseReq) MsgType() MessageType {
	return MessageTypeTreeBaseReq
}

func (d *TreeBaseReq) Equals(other Message) bool {
	cast, ok := other.(*TreeBaseReq)
	if !ok {
		return false
	}

	return d.Name == cast.Name
}

func (d *TreeBaseReq) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		d.Name,
	)
}

func (d *TreeBaseReq) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&d.Name,
	)
}

func (d *TreeBaseReq) Hash() (crypto.Hash, error) {
	return d.HashCacher.Hash(d)
}
