package wire

import (
	"io"

	"fnd/blob"
	"fnd/crypto"

	"fnd.localhost/dwire"
)

type EquivocationProof struct {
	HashCacher

	Name string

	OurEpochHeight     uint16
	OurPayloadPosition uint16
	OurPrevHash        crypto.Hash
	OurReservedRoot    crypto.Hash
	OurPayload         []blob.Sector
	OurSignature       crypto.Signature

	TheirEpochHeight   uint16
	TheirSectorSize    uint16
	TheirSectorTipHash crypto.Hash
	TheirReservedRoot  crypto.Hash
	TheirSignature     crypto.Signature
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
		s.OurEpochHeight == cast.OurEpochHeight &&
		s.OurPayloadPosition == cast.OurPayloadPosition &&
		s.OurPrevHash == cast.OurPrevHash &&
		s.OurReservedRoot == cast.OurReservedRoot &&
		s.OurSignature == cast.OurSignature &&
		s.TheirEpochHeight == cast.TheirEpochHeight &&
		s.TheirSectorSize == cast.TheirSectorSize &&
		s.TheirSectorTipHash == cast.TheirSectorTipHash &&
		s.TheirReservedRoot == cast.TheirReservedRoot &&
		s.TheirSignature == cast.TheirSignature
}

func (s *EquivocationProof) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		s.Name,
		s.OurEpochHeight,
		s.OurPayloadPosition,
		s.OurPrevHash,
		s.OurReservedRoot,
		s.OurPayload,
		s.OurSignature,
		s.TheirEpochHeight,
		s.TheirSectorSize,
		s.TheirSectorTipHash,
		s.TheirReservedRoot,
		s.TheirSignature,
	)
}

func (s *EquivocationProof) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&s.Name,
		&s.OurEpochHeight,
		&s.OurPayloadPosition,
		&s.OurPrevHash,
		&s.OurReservedRoot,
		&s.OurPayload,
		&s.OurSignature,
		&s.TheirEpochHeight,
		&s.TheirSectorSize,
		&s.TheirSectorTipHash,
		&s.TheirReservedRoot,
		&s.TheirSignature,
	)
}

func (s *EquivocationProof) Hash() (crypto.Hash, error) {
	return s.HashCacher.Hash(s)
}
