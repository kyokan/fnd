package blob

import (
	"io"

	"fnd/crypto"

	"golang.org/x/crypto/blake2b"
)

var (
	ZeroHash         crypto.Hash
	ZeroSectorHashes SectorHashes
)

type SectorHashes [SectorCount]crypto.Hash

func (s SectorHashes) Encode(w io.Writer) error {
	for _, h := range s {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}
	}
	return nil
}

func (s *SectorHashes) Decode(r io.Reader) error {
	var res SectorHashes
	var hash crypto.Hash
	for i := 0; i < len(res); i++ {
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}
		res[i] = hash
	}

	*s = res
	return nil
}

func (s SectorHashes) DiffWith(other SectorHashes) int {
	if s == other {
		return -1
	}

	var i int
	for i = 0; i < len(s); i++ {
		if s[i] != other[i] {
			break
		}
	}
	return i
}

func (s SectorHashes) Tip() crypto.Hash {
	var i int
	for i = 0; i < SectorCount; i++ {
		if s[i] == ZeroHash {
			break
		}
	}
	if i == 0 {
		return crypto.ZeroHash
	}
	return s[i-1]
}

func SerialHashSector(sector Sector, prevHash crypto.Hash) crypto.Hash {
	var res crypto.Hash
	hasher, _ := blake2b.New256(nil)
	hasher.Write(prevHash[:])
	hasher.Write(sector[:])
	h := hasher.Sum(nil)
	copy(res[:], h)
	return res
}

// SerialHash returns serial hash of the contents of the reader br
func SerialHash(br io.Reader, prevHash crypto.Hash, sectorSize uint16) (SectorHashes, error) {
	var res SectorHashes
	var sector Sector
	for i := 0; i < int(sectorSize); i++ {
		if _, err := br.Read(sector[:]); err != nil {
			return ZeroSectorHashes, err
		}
		res[i] = SerialHashSector(sector, prevHash)
		prevHash = res[i]
	}
	return res, nil
}
