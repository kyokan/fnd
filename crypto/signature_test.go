package crypto

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSECP256k1Signer(t *testing.T) {
	data, err := hex.DecodeString("86d4da79175bf6984ef62676a20069d35527c45ccc398d46b7fdb9b0783cccf7")
	require.NoError(t, err)
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	require.NoError(t, err)
	hash := Blake2B256([]byte("the quick brown fox jumps over the lazy dog"))
	altHash := Blake2B256([]byte("the quick brown fox jumps over the lazy cat"))
	signer := NewSECP256k1Signer(priv)
	sig, err := signer.Sign(hash)
	require.Equal(t, "1b7ede7d6760bbc5985fbcf54e22f2c34fa3f63c33969c76d349186041044f92493c0b74c900abbb7ffe05f2013032507d9dff09ae25ed9b91eaa91affe2e7cdbb", hex.EncodeToString(sig[:]))
	require.NoError(t, err)
	require.True(t, VerifySigPub(signer.Pub(), sig, hash))
	require.True(t, VerifySigHashedPub(HashPub(signer.Pub()), sig, hash))
	require.False(t, VerifySigPub(signer.Pub(), sig, altHash))
	require.False(t, VerifySigHashedPub(HashPub(signer.Pub()), sig, altHash))
}
