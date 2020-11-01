package primitives

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddress_ToBech32(t *testing.T) {
	hash, err := hex.DecodeString("6d5571fdbca1019cd0f0cd792d1b0bdfa7651c7e")
	require.NoError(t, err)
	addr := &Address{
		Version: 0,
		Hash:    hash,
	}
	bech, err := addr.ToBech32(NetworkMainnet.AddressHRP())
	require.NoError(t, err)
	require.Equal(t, "hs1qd42hrldu5yqee58se4uj6xctm7nk28r70e84vx", bech)
}
