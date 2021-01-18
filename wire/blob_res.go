package wire

import (
	"io"

	"fnd/blob"
	"fnd/crypto"
	"fnd.localhost/dwire"
)

type BlobRes struct {
	HashCacher

	Name            string
	EpochHeight     uint16
	PayloadPosition uint16
	PrevHash        crypto.Hash
	ReservedRoot    crypto.Hash
	Payload         []blob.Sector
}

var _ Message = (*BlobRes)(nil)

func (s *BlobRes) MsgType() MessageType {
	return MessageTypeBlobRes
}

func (s *BlobRes) Equals(other Message) bool {
	cast, ok := other.(*BlobRes)
	if !ok {
		return false
	}

	return s.Name == cast.Name &&
		s.EpochHeight == cast.EpochHeight &&
		s.PayloadPosition == cast.PayloadPosition &&
		s.PrevHash == cast.PrevHash &&
		s.ReservedRoot == cast.ReservedRoot
}

func (s *BlobRes) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		s.Name,
		s.EpochHeight,
		s.PayloadPosition,
		s.PrevHash,
		s.ReservedRoot,
		s.Payload,
	)
}

func (s *BlobRes) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&s.Name,
		&s.EpochHeight,
		&s.PayloadPosition,
		&s.PrevHash,
		&s.ReservedRoot,
		&s.Payload,
	)
}

func (s *BlobRes) Hash() (crypto.Hash, error) {
	return s.HashCacher.Hash(s)
}
