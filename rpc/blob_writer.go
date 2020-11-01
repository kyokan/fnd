package rpc

import (
	"context"
	"fnd/blob"
	"fnd/crypto"
	apiv1 "fnd/rpc/v1"
	"github.com/pkg/errors"
	"io"
	"time"
)

type BlobWriter struct {
	client    apiv1.Footnotev1Client
	signer    crypto.Signer
	name      string
	txID      uint32
	opened    bool
	committed bool
	offset    int64
}

func NewBlobWriter(client apiv1.Footnotev1Client, signer crypto.Signer, name string) *BlobWriter {
	return &BlobWriter{
		client: client,
		signer: signer,
		name:   name,
	}
}

func (b *BlobWriter) Open() error {
	if b.opened {
		panic("writer already open")
	}
	if b.committed {
		panic("writer committed")
	}
	checkoutRes, err := b.client.Checkout(context.Background(), &apiv1.CheckoutReq{
		Name: b.name,
	})
	if err != nil {
		return errors.Wrap(err, "failed to check out blob")
	}
	b.txID = checkoutRes.TxID
	b.opened = true
	return nil
}

func (b *BlobWriter) Truncate() error {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}
	_, err := b.client.Truncate(context.Background(), &apiv1.TruncateReq{
		TxID: b.txID,
	})
	if err != nil {
		return errors.Wrap(err, "error truncating blob")
	}
	return nil
}

func (b *BlobWriter) Seek(offset int64, whence int) (int64, error) {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}

	switch whence {
	case io.SeekStart:
		if b.offset > blob.Size {
			return b.offset, errors.New("seek beyond blob bounds")
		}
		b.offset = offset
	case io.SeekCurrent:
		next := b.offset + offset
		if next > blob.Size {
			return b.offset, errors.New("seek beyond blob bounds")
		}
		b.offset = next
	case io.SeekEnd:
		next := blob.Size - offset
		if next < 0 {
			return b.offset, errors.New("seek beyond blob bounds")
		}
		b.offset = next
	default:
		panic("invalid whence")
	}
	return b.offset, nil
}

func (b *BlobWriter) WriteAt(p []byte, off int64) (int, error) {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}

	var clientErr error
	n := len(p)
	if off+int64(n) > blob.Size {
		clientErr = errors.New("write beyond blob bounds")
		n = blob.Size - int(off)
	}

	res, err := b.client.WriteAt(context.Background(), &apiv1.WriteAtReq{
		TxID:   b.txID,
		Offset: uint32(b.offset),
		Data:   p[:n],
	})
	if err != nil {
		return 0, errors.Wrap(err, "error writing blob")
	}
	if res.WriteErr != "" {
		return int(res.BytesWritten), errors.Wrap(errors.New(res.WriteErr), "error writing blob")
	}
	return n, clientErr
}

func (b *BlobWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	n, err := b.WriteAt(p, b.offset)
	b.offset += int64(n)
	if err != nil {
		return n, errors.Wrap(err, "error writing blob")
	}
	return n, nil
}

func (b *BlobWriter) Commit(broadcast bool) error {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}
	precommitRes, err := b.client.PreCommit(context.Background(), &apiv1.PreCommitReq{
		TxID: b.txID,
	})
	if err != nil {
		return errors.Wrap(err, "error retrieving precommit")
	}
	ts := time.Now()
	var mr crypto.Hash
	copy(mr[:], precommitRes.MerkleRoot)
	sig, err := blob.SignSeal(b.signer, b.name, ts, mr, crypto.ZeroHash)
	if err != nil {
		return errors.Wrap(err, "error sealing blob")
	}
	_, err = b.client.Commit(context.Background(), &apiv1.CommitReq{
		TxID:      b.txID,
		Timestamp: uint64(ts.Unix()),
		Signature: sig[:],
		Broadcast: broadcast,
	})
	if err != nil {
		return errors.Wrap(err, "error sending commit")
	}
	b.committed = true
	return nil
}
