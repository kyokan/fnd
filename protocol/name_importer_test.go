package protocol

import (
	"ddrp/testutil/testcrypto"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseDDRPKeyRecord(t *testing.T) {
	invalid := []string{
		"DDRPKEY:000000000000000000000000000000000000000000000000000000000000000000",
		"DDRPKEY:whateverwhateverwhateverwhateverwhateverwhateverwhateverwhateverwh",
		"DDRPKEY",
		"DDRPKEY:",
		"DDRPKEY:whatever",
		"DDRPKEY",
		"",
		"wibble",
	}
	for _, rec := range invalid {
		_, err := ParseDDRPKeyRecord(rec)
		require.Error(t, err)
	}

	_, expPub := testcrypto.RandKey()
	rec := fmt.Sprintf("DDRPKEY:%x", expPub.SerializeCompressed())
	actPub, err := ParseDDRPKeyRecord(rec)
	require.NoError(t, err)
	require.True(t, expPub.IsEqual(actPub))
}
