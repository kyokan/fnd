package rpc

import (
	"context"
	"errors"
	"fnd/blob"
	apiv1 "fnd/rpc/v1"
	"io"
)

type BlobReader struct {
	client apiv1.DDRPv1Client
	name   string
	off    int64
}

func NewBlobReader(client apiv1.DDRPv1Client, name string) *BlobReader {
	return &BlobReader{
		client: client,
		name:   name,
	}
}

func (b *BlobReader) ReadAt(p []byte, off int64) (int, error) {
	res, err := b.client.ReadAt(context.Background(), &apiv1.ReadAtReq{
		Name:   b.name,
		Offset: uint32(off),
		Len:    uint32(len(p)),
	})
	if err != nil {
		return 0, err
	}
	n := len(res.Data)
	copy(p, res.Data)
	if n != len(p) {
		return n, errors.New("unequal read - should not happen")
	}
	return n, nil
}

func (b *BlobReader) Read(p []byte) (int, error) {
	if b.off == blob.Size {
		return 0, io.EOF
	}
	toRead := len(p)
	if b.off+int64(toRead) > blob.Size {
		toRead = blob.Size - int(b.off)
	}
	n, err := b.ReadAt(p[:toRead], b.off)
	b.off += int64(n)
	return n, err
}

func (b *BlobReader) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if b.off > blob.Size {
			return b.off, errors.New("seek beyond blob bounds")
		}
		b.off = off
	case io.SeekCurrent:
		next := b.off + off
		if next > blob.Size {
			return b.off, errors.New("seek beyond blob bounds")
		}
		b.off = next
	case io.SeekEnd:
		next := blob.Size - off
		if next < 0 {
			return b.off, errors.New("seek beyond blob bounds")
		}
		b.off = next
	default:
		panic("invalid whence")
	}
	return b.off, nil
}
