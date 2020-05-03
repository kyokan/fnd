package blob

import (
	"ddrp/testutil/testfs"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func newTempBlobFile(t *testing.T) (*os.File, func()) {
	f, done := testfs.NewTempFile(t)
	require.NoError(t, f.Truncate(Size))
	return f, done
}
