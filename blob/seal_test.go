package blob

import (
	"encoding/hex"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/testutil/testcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSealHash(t *testing.T) {
	zeroTime := time.Unix(0, 0)
	h := SealHash("testname", zeroTime, crypto.ZeroHash, crypto.ZeroHash)
	assert.Equal(t, "694802db0f9b9b72725cf9c108b7a496bf20e1bdcc4d7feb9c42b8df1a08823b", hex.EncodeToString(h[:]))
}

func TestSignSeal(t *testing.T) {
	zeroTime := time.Unix(0, 0)
	priv, _ := testcrypto.FixedKey(t)
	sig, err := SignSeal(crypto.NewSECP256k1Signer(priv), "testname", zeroTime, crypto.ZeroHash, crypto.ZeroHash)
	require.NoError(t, err)
	assert.Equal(t, "1cdb0e3aa14a5489cc1bfcf25843d6747cb9412d6200c35b69dd5fb9cb133ebc7c339ea9d5bb0cce8a9b20e84642757d9a2bc82e9e0a777a9641dd05fb9e5e4836", sig.String())
}
