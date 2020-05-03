package crypto

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
)

type Signature [65]byte

func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}

func NewSignatureFromBytes(b []byte) (Signature, error) {
	var sig Signature
	if len(b) != 65 {
		return sig, errors.New("signature must be 65 bytes")
	}
	copy(sig[:], b)
	return sig, nil
}

type Signer interface {
	Sign(Hasher) (Signature, error)
	Pub() *btcec.PublicKey
}

type SECP256k1Signer struct {
	pk *btcec.PrivateKey
}

func NewSECP256k1Signer(pk *btcec.PrivateKey) *SECP256k1Signer {
	return &SECP256k1Signer{
		pk: pk,
	}
}

func (s *SECP256k1Signer) Sign(hasher Hasher) (Signature, error) {
	var sig Signature
	hash, err := hasher.Hash()
	if err != nil {
		return sig, err
	}

	sigBuf, err := btcec.SignCompact(btcec.S256(), s.pk, hash.Bytes(), false)
	if err != nil {
		return sig, err
	}
	copy(sig[:], sigBuf)
	return sig, nil
}

func (s *SECP256k1Signer) Pub() *btcec.PublicKey {
	return s.pk.PubKey()
}

func VerifySigPub(pub *btcec.PublicKey, signature Signature, msg Hasher) bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	key, _, err := btcec.RecoverCompact(btcec.S256(), signature[:], hash.Bytes())
	if err != nil {
		return false
	}
	return key.IsEqual(pub)
}

func VerifySigHashedPub(hashedPub [32]byte, signature Signature, msg Hasher) bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	key, _, err := btcec.RecoverCompact(btcec.S256(), signature[:], hash.Bytes())
	if err != nil {
		return false
	}
	idFromKey := HashPub(key)
	return idFromKey == hashedPub
}
