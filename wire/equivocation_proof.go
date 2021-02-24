package wire

import (
	"io"

	"fnd/blob"
	"fnd/crypto"

	"fnd.localhost/dwire"
)

type EquivocationProof struct {
	HashCacher

	Name            string
	EpochHeight     uint16
	PayloadPosition uint16
	PrevHash        crypto.Hash
	ReservedRoot    crypto.Hash
	Payload         []blob.Sector
	Signature       crypto.Signature
}

var _ Message = (*EquivocationProof)(nil)

func (s *EquivocationProof) MsgType() MessageType {
	return MessageTypeEquivocationProof
}

func (s *EquivocationProof) Equals(other Message) bool {
	cast, ok := other.(*EquivocationProof)
	if !ok {
		return false
	}

	return s.Name == cast.Name &&
		s.EpochHeight == cast.EpochHeight &&
		s.PayloadPosition == cast.PayloadPosition &&
		s.PrevHash == cast.PrevHash &&
		s.ReservedRoot == cast.ReservedRoot &&
		s.Signature == cast.Signature
}

func (s *EquivocationProof) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		s.Name,
		s.EpochHeight,
		s.PayloadPosition,
		s.PrevHash,
		s.ReservedRoot,
		s.Payload,
		s.Signature,
	)
}

func (s *EquivocationProof) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&s.Name,
		&s.EpochHeight,
		&s.PayloadPosition,
		&s.PrevHash,
		&s.ReservedRoot,
		&s.Payload,
		&s.Signature,
	)
}

func (s *EquivocationProof) Hash() (crypto.Hash, error) {
	return s.HashCacher.Hash(s)
}
