package protocol

import (
	"fnd/blob"
	"fnd/crypto"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/wire"
	"time"

	"fnd.localhost/handshake/primitives"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DefaultSyncerBlobResTimeout = 15 * time.Second
)

var (
	ErrInvalidPayloadSignature = errors.New("update signature is invalid")
	ErrSyncerNoProgress        = errors.New("sync not progressing")
	ErrSyncerMaxAttempts       = errors.New("reached max sync attempts")
)

type SyncSectorsOpts struct {
	Timeout     time.Duration
	Mux         *p2p.PeerMuxer
	Tx          blob.Transaction
	Peers       *PeerSet
	EpochHeight uint16
	SectorSize  uint16
	PrevHash    crypto.Hash
	Name        string
	DB          *leveldb.DB
}

type payloadRes struct {
	peerID crypto.Hash
	msg    *wire.BlobRes
}

func validateBlobRes(opts *SyncSectorsOpts, name string, epochHeight, sectorSize uint16, mr crypto.Hash, rr crypto.Hash, sig crypto.Signature) error {
	if err := primitives.ValidateName(name); err != nil {
		return errors.Wrap(err, "update name is invalid")
	}
	banned, err := store.NameIsBanned(opts.DB, name)
	if err != nil {
		return errors.Wrap(err, "error reading name ban state")
	}
	if banned {
		return errors.New("name is banned")
	}
	info, err := store.GetNameInfo(opts.DB, name)
	if err != nil {
		return errors.Wrap(err, "error reading name info")
	}
	h := blob.SealHash(name, epochHeight, sectorSize, mr, rr)
	if !crypto.VerifySigPub(info.PublicKey, sig, h) {
		return ErrInvalidPayloadSignature
	}
	return nil
}

func SyncSectors(opts *SyncSectorsOpts) error {
	lgr := log.WithModule("payload-syncer").Sub("name", opts.Name)
	errs := make(chan error)
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
				if opts.SectorSize != msg.PayloadPosition {
					lgr.Trace("received unexpected payload position", "sector_size", opts.SectorSize, "payload_position", msg.PayloadPosition)
					continue
				}
				sectorSize := msg.PayloadPosition + uint16(len(msg.Payload))
				if int(sectorSize) > blob.Size {
					lgr.Trace("received unexpected payload size", "sector_size", sectorSize)
					continue
				}
				var sectorTipHash crypto.Hash = opts.PrevHash
				for i := 0; int(i) < len(msg.Payload); i++ {
					sectorTipHash = blob.SerialHashSector(msg.Payload[i], sectorTipHash)
				}
				if err := validateBlobRes(opts, msg.Name, msg.EpochHeight, sectorSize, sectorTipHash, msg.ReservedRoot, msg.Signature); err != nil {
					lgr.Trace("blob res validation failed", "err", err)
					errs <- err
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

	var err error
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
		case err = <-errs:
			lgr.Warn("payload syncing failed")
			break payloadLoop
		}
	}

	unsubRes()
	close(doneCh)
	return err
}
