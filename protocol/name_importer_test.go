package protocol

import (
	"encoding/base64"
	"fmt"
	"fnd/testutil/testcrypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseFNRecord(t *testing.T) {
	invalid := []string{
		"f000000000000000000000000000000000000000000000000000000000000000000",
		"fwhateverwhateverwhateverwhateverwhateverwhateverwhateverwhateverwh",
		"f",
		"f",
		"fwhatever",
		"f",
		"",
		"wibble",
	}
	for _, rec := range invalid {
		_, err := ParseFNRecord(rec)
		require.Error(t, err)
	}

	_, expPub := testcrypto.RandKey()
	rec := fmt.Sprintf("f%v", base64.StdEncoding.EncodeToString(expPub.SerializeCompressed()))
	actPub, err := ParseFNRecord(rec)
	require.NoError(t, err)
	require.True(t, expPub.IsEqual(actPub))
}
