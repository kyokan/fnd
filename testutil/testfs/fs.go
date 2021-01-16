package testfs

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func NewTempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "ddrptest_")
	require.NoError(t, err)
	return dir, func() {
		require.NoError(t, os.RemoveAll(dir))
	}
}

func NewTempFile(t *testing.T) (*os.File, func()) {
	f, err := ioutil.TempFile("", "ddrptest_")
	require.NoError(t, err)
	return f, func() {
		require.NoError(t, f.Close())
		require.NoError(t, os.Remove(f.Name()))
	}
}
