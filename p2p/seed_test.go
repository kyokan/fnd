package p2p

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLookupDNSSeeds(t *testing.T) {
	seeds, err := ResolveDNSSeeds("seeds-test.merkleblock.com")
	require.NoError(t, err)

	require.Equal(t, 1, len(seeds))
	require.Contains(t, seeds, "78.46.17.17")
}
