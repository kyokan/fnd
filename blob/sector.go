package blob

import "io"

type Sector [SectorBytes]byte

var ZeroSector Sector

func (s Sector) Encode(w io.Writer) error {
	_, err := w.Write(s[:])
	return err
}

func (s *Sector) Decode(r io.Reader) error {
	var newSector Sector
	if _, err := io.ReadFull(r, newSector[:]); err != nil {
		return err
	}
	*s = newSector
	return nil
}
