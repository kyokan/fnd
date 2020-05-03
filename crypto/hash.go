package crypto

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/blake2b"
	"io"
)

type Hash [32]byte

var ZeroHash Hash

type Hasher interface {
	Hash() (Hash, error)
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Encode(w io.Writer) error {
	_, err := w.Write(h.Bytes())
	return err
}

func (h *Hash) Decode(r io.Reader) error {
	var buf Hash
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return err
	}
	*h = buf
	return nil
}

func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%x\"", h[:])), nil
}

func (h *Hash) UnmarshalJSON(b []byte) error {
	var peerIDStr string
	if err := json.Unmarshal(b, &peerIDStr); err != nil {
		return err
	}
	hash, err := NewHashFromHex(peerIDStr)
	if err != nil {
		return err
	}
	*h = hash
	return nil
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) Hash() (Hash, error) {
	return h, nil
}

func Blake2B256(data ...[]byte) Hash {
	// never returns an error if key is nil
	h, _ := blake2b.New256(nil)
	for _, chunk := range data {
		h.Write(chunk)
	}
	b := h.Sum(nil)
	var out Hash
	copy(out[:], b)
	return out
}

func HashPub(pub *btcec.PublicKey) Hash {
	return Blake2B256(pub.SerializeCompressed())
}

func NewHashFromBytes(b []byte) (Hash, error) {
	if len(b) != 32 {
		return ZeroHash, errors.New("hash must be 32 bytes")
	}
	var h Hash
	copy(h[:], b)
	return h, nil
}

func NewHashFromHex(in string) (Hash, error) {
	b, err := hex.DecodeString(in)
	if err != nil {
		return ZeroHash, nil
	}
	return NewHashFromBytes(b)
}
