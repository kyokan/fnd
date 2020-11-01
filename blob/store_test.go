package blob

import (
	"crypto/rand"
	"fnd/testutil/testfs"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
	"io"
	"os"
	"path"
	"testing"
)

func TestBlobStore(t *testing.T) {
	dir, done := testfs.NewTempDir(t)
	defer done()

	_, err := os.Stat(path.Join(dir, "f2c9/fa97/ffbb/75c2/c523/a8d6/839c/7b41/2321/2ce2/cc03/50d2/7037/3059/646b/86ff_blob"))
	require.True(t, os.IsNotExist(err))

	store := NewStore(dir)
	blob, err := store.Open("fooname")
	require.NoError(t, err)

	blobFile := path.Join(dir, "f2c9/fa97/ffbb/75c2/c523/a8d6/839c/7b41/2321/2ce2/cc03/50d2/7037/3059/646b/86ff_blob")
	info, err := os.Stat(blobFile)
	require.NoError(t, err)
	require.False(t, info.IsDir())
	require.NoError(t, blob.Close())

	f, err := os.OpenFile(blobFile, os.O_RDWR, 0755)
	require.NoError(t, err)
	h, _ := blake2b.New256(nil)
	_, err = io.CopyN(f, io.TeeReader(rand.Reader, h), Size)
	require.NoError(t, err)
	hash := h.Sum(nil)
	h.Reset()

	blob, err = store.Open("fooname")
	require.NoError(t, err)
	_, err = io.Copy(h, NewReader(blob))
	require.NoError(t, err)
	require.Equal(t, hash, h.Sum(nil))
}
