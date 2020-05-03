package testcrypto

import (
	"crypto/rand"
	"ddrp/crypto"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
	"testing"
)

func RandKey() (*btcec.PrivateKey, *btcec.PublicKey) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}
	return priv, priv.PubKey()
}

func FixedKey(t *testing.T) (*btcec.PrivateKey, *btcec.PublicKey) {
	data, err := hex.DecodeString("86d4da79175bf6984ef62676a20069d35527c45ccc398d46b7fdb9b0783cccf7")
	require.NoError(t, err)
	return btcec.PrivKeyFromBytes(btcec.S256(), data)
}

func FixedSigner(t *testing.T) crypto.Signer {
	priv, _ := FixedKey(t)
	return crypto.NewSECP256k1Signer(priv)
}

type RandomSigner struct {
}

func NewRandomSigner() crypto.Signer {
	return &RandomSigner{}
}

func (r RandomSigner) Sign(crypto.Hasher) (crypto.Signature, error) {
	var sig crypto.Signature
	_, err := rand.Read(sig[:])
	if err != nil {
		panic(err)
	}
	return sig, nil
}

func (r RandomSigner) Pub() *btcec.PublicKey {
	_, pub := RandKey()
	return pub
}
