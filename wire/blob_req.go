package wire

import (
	"io"

	"fnd/crypto"
	"fnd.localhost/dwire"
)

type BlobReq struct {
	HashCacher

	Name        string
	EpochHeight uint16
	SectorSize  uint16
}

var _ Message = (*BlobReq)(nil)

func (u *BlobReq) MsgType() MessageType {
	return MessageTypeBlobReq
}

func (u *BlobReq) Equals(other Message) bool {
	cast, ok := other.(*BlobReq)
	if !ok {
		return false
	}

	return u.Name == cast.Name &&
		u.EpochHeight == cast.EpochHeight &&
		u.SectorSize == cast.SectorSize
}

func (u *BlobReq) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		u.Name,
		u.EpochHeight,
		u.SectorSize,
	)
}

func (u *BlobReq) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&u.Name,
		&u.EpochHeight,
		&u.SectorSize,
	)
}

func (u *BlobReq) Hash() (crypto.Hash, error) {
	return u.HashCacher.Hash(u)
}
