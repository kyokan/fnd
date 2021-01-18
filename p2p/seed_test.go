package p2p

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLookupDNSSeeds(t *testing.T) {
	seeds, err := ResolveDNSSeeds("seeds-test.fnd.network")
	require.NoError(t, err)

	require.Equal(t, 2, len(seeds))
	require.Contains(t, seeds, "10.1.0.1")
	require.Contains(t, seeds, "10.1.0.2")
}
