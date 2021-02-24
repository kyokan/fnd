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
	ErrPayloadEquivocation     = errors.New("update payload is equivocated")
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
				// Verify that the remote is at the same epoch height or lower
				if opts.EpochHeight > msg.EpochHeight {
					lgr.Trace("received unexpected epoch height", "expected_epoch_height", opts.EpochHeight, "received_epoch_height", msg.EpochHeight)
					continue
				}
				// Verify that we received the payload starting from the sector
				// we requested in blob request.  opts.SectorSize contains our
				// current known sector size, which is what we send in blob
				// request.
				if opts.SectorSize != msg.PayloadPosition {
					lgr.Trace("received unexpected payload position", "sector_size", opts.SectorSize, "payload_position", msg.PayloadPosition)
					continue
				}
				sectorSize := msg.PayloadPosition + uint16(len(msg.Payload))
				// Additional sanity check: make sure that update does not overflow max sectors.
				if int(sectorSize) > blob.MaxSectors {
					lgr.Trace("received unexpected sector size", "sector_size", sectorSize, "max", blob.MaxSectors)
					continue
				}
				// Generate the current tip hash from prev hash and the payload
				// sectors.
				var sectorTipHash crypto.Hash = opts.PrevHash
				for i := 0; int(i) < len(msg.Payload); i++ {
					sectorTipHash = blob.SerialHashSector(msg.Payload[i], sectorTipHash)
				}
				// Verify that the update is valid by using the recomputed
				// sector size, sector tip hash and other metadata. This data
				// is first hashed and the signature is validated against the
				// name's pubkey. See validateBlobRes.
				// TODO: store the latest tip hash
				if err := validateBlobRes(opts, msg.Name, msg.EpochHeight, sectorSize, sectorTipHash, msg.ReservedRoot, msg.Signature); err != nil {
					lgr.Trace("blob res validation failed", "err", err)
					// Verify that the prev hash from the remote matches our
					// current tip hash i.e.  the update starts _after_ our
					// latest sector and both the sector hashes match.  A
					// mismatch indicates a proof of equivocation.
					if opts.PrevHash != msg.PrevHash {
						lgr.Trace("received unexpected prev hash", "expected_prev_hash", opts.PrevHash, "received_prev_hash", msg.PrevHash)
						if err := validateBlobRes(opts, msg.Name, msg.EpochHeight, sectorSize, msg.PrevHash, msg.ReservedRoot, msg.Signature); err != nil {
							if opts.EpochHeight == msg.EpochHeight {
								if err := store.WithTx(opts.DB, func(tx *leveldb.Transaction) error {
									return store.SetEquivocationProofTx(tx, msg.Name, msg)
								}); err != nil {
									lgr.Trace("error writing equivocation proof", "err", err)
								}
								// TODO: handle equivocation proof
								// local
								// -> update { sectorSize: 0 }
								// <- BlobReq { sectorSize 0xff }
								// -> BlobRes { equivocation proof }
								// remote
								// <- update { sectorSize: 0 }
								// -> BlobReq { sectorSize 0xff }
								// <- BlobRes { equivocation proof }
								update := &wire.Update{
									Name:        msg.Name,
									EpochHeight: msg.EpochHeight,
									SectorSize:  0,
								}
								p2p.GossipAll(opts.Mux, update)
							}
							errs <- ErrPayloadEquivocation
							break
						}
					}
					// If prev hash matches, we have an invalid signature,
					// which cannot be used as a proof of equivocation.
					// TODO: ban the peer as it is clearly sending invalid data
					errs <- errors.Wrap(ErrInvalidPayloadSignature, "signature validation failed")
					break
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
