package rpc

import (
	"context"
	"ddrp/blob"
	"ddrp/crypto"
	apiv1 "ddrp/rpc/v1"
	"github.com/pkg/errors"
	"io"
	"time"
)

const (
	BlobWriterMaxChunkSize = 1 * 1024 * 1024
)

type BlobWriter struct {
	Client      apiv1.DDRPv1Client
	Signer      crypto.Signer
	Name        string
	Timeout     time.Duration
	StartOffset int
	Truncate    bool
	Broadcast   bool
	writeClient apiv1.DDRPv1_WriteClient
	txID        uint32
	offset      int
}

func NewBlobWriter(client apiv1.DDRPv1Client, signer crypto.Signer, name string) *BlobWriter {
	return &BlobWriter{
		Client:    client,
		Signer:    signer,
		Name:      name,
		Broadcast: true,
	}
}

func (b *BlobWriter) Write(p []byte) (int, error) {
	if b.writeClient == nil {
		ctx := context.Background()
		checkoutRes, err := b.Client.Checkout(ctx, &apiv1.CheckoutReq{
			Name: b.Name,
		})
		if err != nil {
			return 0, errors.Wrap(err, "failed to check out blob")
		}

		if b.Truncate {
			_, err = b.Client.Truncate(ctx, &apiv1.TruncateReq{
				TxID: checkoutRes.TxID,
			})
			if err != nil {
				return 0, errors.Wrap(err, "failed to truncate blob")
			}
		}

		b.txID = checkoutRes.TxID
		wc, err := b.Client.Write(context.Background())
		if err != nil {
			return 0, errors.Wrap(err, "failed to open write stream")
		}
		b.writeClient = wc
		b.offset = b.StartOffset
	}

	toWrite := len(p)
	if toWrite == 0 {
		return 0, nil
	}

	var writeErr error
	remaining := blob.Size - b.offset
	if toWrite > remaining {
		writeErr = io.EOF
		toWrite = remaining
	}
	if toWrite > BlobWriterMaxChunkSize {
		writeErr = errors.New("chunk size too large")
		toWrite = BlobWriterMaxChunkSize
	}

	err := b.writeClient.Send(&apiv1.WriteReq{
		TxID:   b.txID,
		Offset: uint32(b.offset),
		Data:   p[:toWrite],
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to send write request")
	}

	b.offset += toWrite
	return toWrite, writeErr
}

func (b *BlobWriter) Close() error {
	if b.writeClient == nil {
		return nil
	}

	if _, err := b.writeClient.CloseAndRecv(); err != nil {
		return errors.Wrap(err, "failed to close write stream")
	}
	ctx := context.Background()
	precommitRes, err := b.Client.PreCommit(ctx, &apiv1.PreCommitReq{
		TxID: b.txID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to perform precommit")
	}
	ts := time.Now()
	var mr crypto.Hash
	copy(mr[:], precommitRes.MerkleRoot)
	sig, err := blob.SignSeal(b.Signer, b.Name, ts, mr, crypto.ZeroHash)
	if err != nil {
		return errors.Wrap(err, "failed to sign merkle root")
	}
	_, err = b.Client.Commit(ctx, &apiv1.CommitReq{
		TxID:      b.txID,
		Timestamp: uint64(ts.Unix()),
		Signature: sig[:],
		Broadcast: b.Broadcast,
	})
	if err != nil {
		return errors.Wrap(err, "failed to commit blob")
	}
	return nil
}
