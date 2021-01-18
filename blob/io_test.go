package blob

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type zeroReader struct{}

func (z *zeroReader) Read(p []byte) (int, error) {
	for i := 0; i < len(p); i++ {
		p[i] = 0x00
	}
	return len(p), nil
}

func (z *zeroReader) ReadAt(p []byte, off int64) (int, error) {
	return z.Read(p)
}

type readerWrapper struct {
	R   io.ReaderAt
	Off int64
}

func (r *readerWrapper) ReadAt(p []byte, off int64) (n int, err error) {
	r.Off = off
	return r.R.ReadAt(p, off)
}

type discardWriter struct{}

func (d *discardWriter) WriteAt(p []byte, off int64) (n int, err error) {
	return len(p), nil
}

type writerWrapper struct {
	W   io.WriterAt
	Off int64
}

func (w *writerWrapper) WriteAt(p []byte, off int64) (n int, err error) {
	w.Off = off
	return w.W.WriteAt(p, off)
}

func TestReadBlobAt(t *testing.T) {
	tests := []struct {
		name string
		len  int
		off  int64
		n    int
		err  error
	}{
		{
			"offset past blob bounds",
			0,
			Size + 1,
			0,
			io.EOF,
		},
		{
			"len + offset past blob bounds",
			10,
			Size - 1,
			1,
			io.EOF,
		},
		{
			"len + offset past blob bounds",
			Size + 1,
			0,
			Size,
			io.EOF,
		},
		{
			"len + offset past blob bounds",
			Size + 1,
			10,
			Size - 10,
			io.EOF,
		},
		{
			"len + offset within blob bounds",
			10,
			10,
			10,
			nil,
		},
	}
	zr := new(zeroReader)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := ReadBlobAt(zr, make([]byte, tt.len), tt.off)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.n, n)
		})
	}
}

func TestReadSector(t *testing.T) {
	tests := []struct {
		id     uint8
		offset int64
	}{
		{
			0,
			0,
		},
		{
			255,
			16711680,
		},
	}
	r := &readerWrapper{
		R: new(zeroReader),
	}
	for _, tt := range tests {
		sector, err := ReadSector(r, tt.id)
		require.NoError(t, err)
		require.Equal(t, ZeroSector, sector)
		require.Equal(t, tt.offset, r.Off)
	}
}

func TestWriteBlobAt(t *testing.T) {
	tests := []struct {
		name string
		len  int
		off  int64
		n    int
		err  error
	}{
		{
			"offset past blob bounds",
			0,
			Size + 1,
			0,
			ErrWriteBeyondBlobBounds,
		},
		{
			"len + offset past blob bounds",
			10,
			Size - 1,
			1,
			ErrWriteBeyondBlobBounds,
		},
		{
			"len + offset past blob bounds",
			Size + 1,
			0,
			Size,
			ErrWriteBeyondBlobBounds,
		},
		{
			"len + offset past blob bounds",
			Size + 1,
			10,
			Size - 10,
			ErrWriteBeyondBlobBounds,
		},
		{
			"len + offset within blob bounds",
			10,
			10,
			10,
			nil,
		},
	}
	w := new(discardWriter)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := WriteBlobAt(w, make([]byte, tt.len), tt.off)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.n, n)
		})
	}
}

func TestWriteSector(t *testing.T) {
	tests := []struct {
		id     uint16
		offset int64
	}{
		{
			0,
			0,
		},
		{
			255,
			16711680,
		},
	}
	w := &writerWrapper{
		W: new(discardWriter),
	}
	for _, tt := range tests {
		err := WriteSector(w, tt.id, ZeroSector)
		require.NoError(t, err)
		require.Equal(t, tt.offset, w.Off)
	}
}
