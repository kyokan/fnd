package wire

import (
	"fnd/crypto"
	"fnd.localhost/dwire"
	"io"
)

type SectorReq struct {
	HashCacher

	Name     string
	SectorID uint8
}

var _ Message = (*SectorReq)(nil)

func (s *SectorReq) MsgType() MessageType {
	return MessageTypeSectorReq
}

func (s *SectorReq) Equals(other Message) bool {
	cast, ok := other.(*SectorReq)
	if !ok {
		return false
	}

	return s.Name == cast.Name &&
		s.SectorID == cast.SectorID
}

func (s *SectorReq) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		s.Name,
		s.SectorID,
	)
}

func (s *SectorReq) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&s.Name,
		&s.SectorID,
	)
}

func (s *SectorReq) Hash() (crypto.Hash, error) {
	return s.HashCacher.Hash(s)
}
