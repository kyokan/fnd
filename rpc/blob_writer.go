package rpc

import (
	"context"
	"fnd/blob"
	"fnd/crypto"
	apiv1 "fnd/rpc/v1"
	"time"

	"github.com/pkg/errors"
)

type BlobWriter struct {
	client        apiv1.Footnotev1Client
	signer        crypto.Signer
	name          string
	epochHeight   uint16
	sectorSize    uint16
	sectorTipHash crypto.Hash
	txID          uint32
	opened        bool
	committed     bool
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
	b.epochHeight = uint16(checkoutRes.EpochHeight)
	b.sectorSize = uint16(checkoutRes.SectorSize)
	sectorTipHash, err := crypto.NewHashFromBytes(checkoutRes.SectorTipHash)
	if err != nil {
		return errors.Wrap(err, "failed to check out blob")
	}
	b.sectorTipHash = sectorTipHash
	b.opened = true
	return nil
}

func (b *BlobWriter) WriteSector(p []byte) (crypto.Hash, error) {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}
	var sector blob.Sector
	copy(sector[:], p)

	res, err := b.client.WriteSector(context.Background(), &apiv1.WriteSectorReq{
		TxID: b.txID,
		Data: p,
	})
	if err != nil {
		return blob.ZeroHash, errors.Wrap(err, "error writing blob")
	}
	if res.WriteErr != "" {
		return blob.ZeroHash, errors.Wrap(errors.New(res.WriteErr), "error writing blob")
	}

	b.sectorSize++
	b.sectorTipHash = blob.SerialHashSector(sector, b.sectorTipHash)

	return b.sectorTipHash, nil
}

func (b *BlobWriter) Reset() error {
	_, err := b.client.ResetEpoch(context.Background(), &apiv1.ResetEpochReq{
		TxID: b.txID,
	})
	b.opened = false
	return err
}

func (b *BlobWriter) Commit(broadcast bool) (crypto.Hash, error) {
	if !b.opened {
		panic("writer not open")
	}
	if b.committed {
		panic("writer committed")
	}
	ts := time.Now()

	sig, err := blob.SignSeal(b.signer, b.name, b.epochHeight, b.sectorSize, b.sectorTipHash, crypto.ZeroHash)
	if err != nil {
		return blob.ZeroHash, errors.Wrap(err, "error sealing blob")
	}
	_, err = b.client.Commit(context.Background(), &apiv1.CommitReq{
		TxID:          b.txID,
		Timestamp:     uint64(ts.Unix()),
		Signature:     sig[:],
		Broadcast:     broadcast,
		EpochHeight:   uint32(b.epochHeight),
		SectorSize:    uint32(b.sectorSize),
		SectorTipHash: b.sectorTipHash.Bytes(),
	})
	if err != nil {
		return blob.ZeroHash, errors.Wrap(err, "error sending commit")
	}
	b.committed = true
	return b.sectorTipHash, nil
}
