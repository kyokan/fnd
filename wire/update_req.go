package wire

import (
	"fnd/crypto"
	"fnd.localhost/dwire"
	"io"
)

type UpdateReq struct {
	HashCacher

	Name        string
	EpochHeight uint16
	SectorSize  uint16
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
		n.EpochHeight == cast.EpochHeight &&
		n.SectorSize == cast.SectorSize
}

func (n *UpdateReq) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		n.Name,
		n.EpochHeight,
		n.SectorSize,
	)
}

func (n *UpdateReq) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&n.Name,
		&n.EpochHeight,
		&n.SectorSize,
	)
}

func (n *UpdateReq) Hash() (crypto.Hash, error) {
	return n.HashCacher.Hash(n)
}
