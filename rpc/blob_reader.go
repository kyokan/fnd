package rpc

import (
	"context"
	"ddrp/blob"
	apiv1 "ddrp/rpc/v1"
	"errors"
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
	if b.off+int64(len(p)) > blob.Size {
		toRead = blob.Size - toRead
	}
	n, err := b.ReadAt(p[:toRead], b.off)
	b.off += int64(n)
	return n, err
}
