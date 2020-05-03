package wire

import (
	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
)

type TreeBaseRes struct {
	HashCacher

	Name       string
	MerkleBase blob.MerkleBase
}

func (d *TreeBaseRes) MsgType() MessageType {
	return MessageTypeTreeBaseRes
}

func (d *TreeBaseRes) Equals(other Message) bool {
	cast, ok := other.(*TreeBaseRes)
	if !ok {
		return false
	}

	return d.Name == cast.Name &&
		d.MerkleBase == cast.MerkleBase
}

func (d *TreeBaseRes) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		d.Name,
		d.MerkleBase,
	)
}

func (d *TreeBaseRes) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&d.Name,
		&d.MerkleBase,
	)
}

func (d *TreeBaseRes) Hash() (crypto.Hash, error) {
	return d.HashCacher.Hash(d)
}
