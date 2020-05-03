package blob

import (
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/dwire"
	"golang.org/x/crypto/blake2b"
	"time"
)

func SealHash(name string, ts time.Time, merkleRoot crypto.Hash, reservedRoot crypto.Hash) crypto.Hash {
	h, _ := blake2b.New256(nil)
	if _, err := h.Write([]byte("DDRPBLOB")); err != nil {
		panic(err)
	}
	if err := dwire.EncodeField(h, name); err != nil {
		panic(err)
	}
	if err := dwire.EncodeField(h, ts); err != nil {
		panic(err)
	}
	if _, err := h.Write(merkleRoot[:]); err != nil {
		panic(err)
	}
	if _, err := h.Write(reservedRoot[:]); err != nil {
		panic(err)
	}

	var out crypto.Hash
	copy(out[:], h.Sum(nil))
	return out
}

func SignSeal(signer crypto.Signer, name string, ts time.Time, merkleRoot crypto.Hash, reservedRoot crypto.Hash) (crypto.Signature, error) {
	h := SealHash(name, ts, merkleRoot, reservedRoot)
	return signer.Sign(h)
}
