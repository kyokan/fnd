package blob

import (
	"encoding/hex"
	"fnd/crypto"
	"fnd/testutil/testcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSealHash(t *testing.T) {
	epochZero := uint16(0)
	sectorSize := uint16(0)
	h := SealHash("testname", epochZero, sectorSize, crypto.ZeroHash, crypto.ZeroHash)
	assert.Equal(t, "ec3f68febf79bfe9439b37c9ae707a7e16abf02e67d1ee29fd530598571e16de", hex.EncodeToString(h[:]))
}

func TestSignSeal(t *testing.T) {
	epochZero := uint16(0)
	sectorSize := uint16(0)
	priv, _ := testcrypto.FixedKey(t)
	sig, err := SignSeal(crypto.NewSECP256k1Signer(priv), "testname", epochZero, sectorSize, crypto.ZeroHash, crypto.ZeroHash)
	require.NoError(t, err)
	assert.Equal(t, "1c3f284e48f72666b5631c707f53c81cc6ba79f3b92d0ae7785189f1ebf43a8174572d16e6ef38de65eb918f6468d7c0bcc04a269fc7346b9da2f29c4b3a08a695", sig.String())
}
