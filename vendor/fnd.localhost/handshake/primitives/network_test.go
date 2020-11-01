package primitives

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNetwork_RPCPort(t *testing.T) {
	require.Equal(t, 12037, NetworkMainnet.RPCPort())
	require.Equal(t, 13037, NetworkTestnet.RPCPort())
	require.Equal(t, 14037, NetworkRegtest.RPCPort())
	require.Equal(t, 15037, NetworkSimnet.RPCPort())
	require.Panics(t, func() {
		Network("foobar").RPCPort()
	})
}
