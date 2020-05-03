package blob

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPool(t *testing.T) {
	var blobs []*BlobMock
	pool := NewPool(func(name string) (Blob, error) {
		blob := new(BlobMock)
		blob.On("Name").Return(name)
		blob.On("Close").Return(nil)
		blobs = append(blobs, blob)
		return blob, nil
	})

	blob1, err := pool.Get("foo")
	require.NoError(t, err)
	blob2, err := pool.Get("foo")
	require.NoError(t, err)
	require.Equal(t, blob1, blob2)
	blob3, err := pool.Get("bar")
	require.NoError(t, err)
	require.NotEqual(t, blob1, blob3)

	require.NoError(t, pool.Put(blob1))
	require.NoError(t, pool.Put(blob2))
	require.NoError(t, pool.Put(blob3))

	require.Equal(t, 0, len(pool.blobs))
	for _, blob := range blobs {
		blob.AssertExpectations(t)
	}

	require.Panics(t, func() {
		pool.Put(new(BlobMock))
	})
}
