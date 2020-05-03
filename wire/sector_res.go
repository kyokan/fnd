package wire

import (
	"ddrp/blob"
	"ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"io"
)

type SectorRes struct {
	HashCacher

	Name     string
	SectorID uint8
	Sector   blob.Sector
}

var _ Message = (*SectorRes)(nil)

func (s *SectorRes) MsgType() MessageType {
	return MessageTypeSectorRes
}

func (s *SectorRes) Equals(other Message) bool {
	cast, ok := other.(*SectorRes)
	if !ok {
		return false
	}

	return s.Name == cast.Name &&
		s.SectorID == cast.SectorID &&
		s.Sector == cast.Sector
}

func (s *SectorRes) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		s.Name,
		s.SectorID,
		s.Sector,
	)
}

func (s *SectorRes) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&s.Name,
		&s.SectorID,
		&s.Sector,
	)
}

func (s *SectorRes) Hash() (crypto.Hash, error) {
	return s.HashCacher.Hash(s)
}
