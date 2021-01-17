package blob

import (
	"fnd/crypto"
	"fnd/dwire"

	"golang.org/x/crypto/blake2b"
)

func SealHash(name string, epochHeight, sectorSize uint16, sectorTipHash crypto.Hash, reservedRoot crypto.Hash) crypto.Hash {
	h, _ := blake2b.New256(nil)
	if _, err := h.Write([]byte("DDRPBLOB")); err != nil {
		panic(err)
	}
	if err := dwire.EncodeField(h, name); err != nil {
		panic(err)
	}
	if err := dwire.EncodeField(h, epochHeight); err != nil {
		panic(err)
	}
	if err := dwire.EncodeField(h, sectorSize); err != nil {
		panic(err)
	}
	if _, err := h.Write(sectorTipHash[:]); err != nil {
		panic(err)
	}
	if _, err := h.Write(reservedRoot[:]); err != nil {
		panic(err)
	}

	var out crypto.Hash
	copy(out[:], h.Sum(nil))
	return out
}

func SignSeal(signer crypto.Signer, name string, epochHeight, sectorSize uint16, sectorTipHash crypto.Hash, reservedRoot crypto.Hash) (crypto.Signature, error) {
	h := SealHash(name, epochHeight, sectorSize, sectorTipHash, reservedRoot)
	return signer.Sign(h)
}
