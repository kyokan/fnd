package protocol

import (
	"fnd/blob"
	"fnd/crypto"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/wire"
	"time"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DefaultSyncerBlobResTimeout = 15 * time.Second
)

var (
	ErrSyncerNoProgress  = errors.New("sync not progressing")
	ErrSyncerMaxAttempts = errors.New("reached max sync attempts")
)

type SyncSectorsOpts struct {
	Timeout       time.Duration
	Mux           *p2p.PeerMuxer
	Tx            blob.Transaction
	Peers         *PeerSet
	EpochHeight   uint16
	SectorSize    uint16
	PrevHash      crypto.Hash
	SectorTipHash crypto.Hash
	Name          string
	DB            *leveldb.DB
}

type payloadRes struct {
	peerID crypto.Hash
	msg    *wire.BlobRes
}

func SyncSectors(opts *SyncSectorsOpts) error {
	lgr := log.WithModule("payload-syncer").Sub("name", opts.Name)
	payloadResCh := make(chan *payloadRes)
	payloadProcessedCh := make(chan struct{}, 1)
	doneCh := make(chan struct{})
	unsubRes := opts.Mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeBlobRes, func(peerID crypto.Hash, envelope *wire.Envelope) {
		payloadResCh <- &payloadRes{
			peerID: peerID,
			msg:    envelope.Message.(*wire.BlobRes),
		}
	}))

	go func() {
		receivedPayloads := make(map[uint16]bool)
		for {
			iter := opts.Peers.Iterator()
			var sendCount int
			for {
				peerID, ok := iter()
				if !ok {
					break
				}
				if sendCount == 7 {
					break
				}
				err := opts.Mux.Send(peerID, &wire.BlobReq{
					Name:        opts.Name,
					EpochHeight: opts.EpochHeight,
					SectorSize:  opts.SectorSize,
				})
				if err != nil {
					lgr.Warn("error fetching payload from peer, trying another", "peer_id", peerID, "err", err)
					continue
				}
				lgr.Debug(
					"requested payload from peer",
					"peer_id", peerID,
				)
				sendCount++
			}
			select {
			case res := <-payloadResCh:
				msg := res.msg
				peerID := res.peerID
				if msg.Name != opts.Name {
					lgr.Trace("received payload for extraneous name", "other_name", msg.Name)
					continue
				}
				if receivedPayloads[msg.PayloadPosition] {
					lgr.Trace("already processed this payload", "payload_position", msg.PayloadPosition, "peer_id", peerID)
					continue
				}
				// TODO: if payloadposition = 0xff, handle equivocation proof
				if opts.SectorSize != msg.PayloadPosition {
					lgr.Trace("received unexpected payload position", "payload_size", opts.SectorSize, "payload_position", msg.PayloadPosition)
					continue
				}
				var sectorTipHash crypto.Hash = opts.PrevHash
				for i := 0; int(i) < len(msg.Payload); i++ {
					sectorTipHash = blob.SerialHashSector(msg.Payload[i], sectorTipHash)
				}
				if sectorTipHash != opts.SectorTipHash {
					lgr.Trace("payload tip hash mismatch", "payload_tip_hash", sectorTipHash, "expected_payload_tip_hash", opts.SectorTipHash)
					err := store.WithTx(opts.DB, func(tx *leveldb.Transaction) error {
						return store.SetEquivocationProofTx(tx, msg.Name, &wire.EquivocationProof{})
					})
					if err != nil {
						lgr.Error("failed to write equivocation proof", "err", err)
					}
					return
				}
				for i := 0; int(i) < len(msg.Payload); i++ {
					if err := opts.Tx.WriteSector(msg.Payload[i]); err != nil {
						lgr.Error("failed to write payload", "payload_id", i, "err", err)
						continue
					}
				}
				receivedPayloads[msg.PayloadPosition] = true
				payloadProcessedCh <- struct{}{}
			case <-doneCh:
				return
			}
		}
	}()

	timeout := time.NewTimer(opts.Timeout)
payloadLoop:
	for {
		lgr.Debug("requesting payload")
		select {
		case <-payloadProcessedCh:
			lgr.Debug("payload processed")
			break payloadLoop
		case <-timeout.C:
			lgr.Warn("payload request timed out")
			break payloadLoop
		}
	}

	unsubRes()
	close(doneCh)
	return nil
}
