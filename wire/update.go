package wire

import (
	"fnd/crypto"
	"io"

	"fnd.localhost/dwire"
)

type Update struct {
	HashCacher

	Name        string
	EpochHeight uint16
	SectorSize  uint16
}

var _ Message = (*Update)(nil)

func (u *Update) MsgType() MessageType {
	return MessageTypeUpdate
}

func (u *Update) Equals(other Message) bool {
	cast, ok := other.(*Update)
	if !ok {
		return false
	}

	return u.Name == cast.Name &&
		u.EpochHeight == cast.EpochHeight &&
		u.SectorSize == cast.SectorSize
}

func (u *Update) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		u.Name,
		u.EpochHeight,
		u.SectorSize,
	)
}

func (u *Update) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&u.Name,
		&u.EpochHeight,
		&u.SectorSize,
	)
}

func (u *Update) Hash() (crypto.Hash, error) {
	return u.HashCacher.Hash(u)
}
