package blob

import (
	"github.com/pkg/errors"
	"io"
)

var (
	ErrWriteBeyondBlobBounds = errors.New("write beyond blob bounds")
)

func ReadBlobAt(r io.ReaderAt, b []byte, off int64) (int, error) {
	if off > Size {
		return 0, io.EOF
	}
	if len(b)+int(off) > Size {
		buf := make([]byte, Size-off)
		n, err := r.ReadAt(buf, off)
		copy(b, buf)
		if err != nil {
			return n, err
		}
		return n, io.EOF
	}
	return r.ReadAt(b, off)
}

func ReadSector(r io.ReaderAt, id uint8) (Sector, error) {
	var sector Sector
	_, err := r.ReadAt(sector[:], int64(id)*int64(SectorBytes))
	return sector, err
}

func WriteBlobAt(w io.WriterAt, b []byte, off int64) (int, error) {
	if off > Size {
		return 0, ErrWriteBeyondBlobBounds
	}
	if len(b)+int(off) > Size {
		buf := make([]byte, Size-off)
		n, err := w.WriteAt(b[:Size-off], off)
		copy(b, buf)
		if err != nil {
			return n, err
		}
		return n, ErrWriteBeyondBlobBounds
	}
	return w.WriteAt(b, off)
}

func WriteSector(w io.WriterAt, id uint16, sector Sector) error {
	_, err := w.WriteAt(sector[:], int64(id)*int64(SectorBytes))
	return err
}

type Reader struct {
	r      io.ReaderAt
	offset int64
}

func NewReader(r io.ReaderAt) io.Reader {
	return &Reader{
		r: r,
	}
}

func (b *Reader) Read(p []byte) (int, error) {
	if b.offset >= Size {
		return 0, io.EOF
	}

	remaining := Size - b.offset
	if remaining < int64(len(p)) {
		buf := make([]byte, remaining, remaining)
		n, err := b.r.ReadAt(buf, b.offset)
		b.offset += int64(n)
		copy(p, buf)
		return n, err
	}

	n, err := b.r.ReadAt(p, b.offset)
	b.offset += int64(n)
	return n, err
}

type Writer struct {
	w      io.WriterAt
	offset int64
}

func NewWriter(w io.WriterAt) io.Writer {
	return &Writer{
		w: w,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	if w.offset >= Size {
		return 0, io.EOF
	}

	remaining := Size - w.offset
	if remaining < int64(len(p)) {
		buf := make([]byte, remaining, remaining)
		n, err := w.w.WriteAt(buf, w.offset)
		w.offset += int64(n)
		copy(p, buf)
		return n, err
	}

	n, err := w.w.WriteAt(p, w.offset)
	w.offset += int64(n)
	return n, err
}
