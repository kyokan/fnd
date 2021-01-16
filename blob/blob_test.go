package blob

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// note: transaction functionality is tested in transaction_test.go
func TestBlob_Reading(t *testing.T) {
	f, done := newTempBlobFile(t)
	defer done()

	_, err := io.CopyN(f, rand.Reader, Size)
	require.NoError(t, err)
	bl := newFromFile("foobar", f)
	require.Equal(t, "foobar", bl.Name())

	var expSector Sector
	_, err = f.ReadAt(expSector[:], 256)
	require.NoError(t, err)
	actSector, err := bl.ReadSector(1)
	require.NoError(t, err)
	require.Equal(t, expSector, actSector)
	actData := make([]byte, 10, 10)
	_, err = bl.ReadAt(actData, 256)
	require.NoError(t, err)
	require.EqualValues(t, expSector[:10], actData)
}
