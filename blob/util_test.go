package blob

import (
	"fnd/testutil/testfs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTempBlobFile(t *testing.T) (*os.File, func()) {
	f, done := testfs.NewTempFile(t)
	require.NoError(t, f.Truncate(Size))
	return f, done
}
