package protocol

import (
	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	DefaultSyncerTreeBaseResTimeout = 10 * time.Second
	DefaultSyncerSectorResTimeout   = 15 * time.Second
)

var (
	ErrNoTreeBaseCandidates = errors.New("no tree base candidates")
	ErrSyncerNoProgress     = errors.New("sync not progressing")
	ErrSyncerMaxAttempts    = errors.New("reached max sync attempts")
)

type SyncTreeBasesOpts struct {
	Timeout    time.Duration
	Mux        *p2p.PeerMuxer
	Peers      *PeerSet
	MerkleRoot crypto.Hash
	Name       string
}

func SyncTreeBases(opts *SyncTreeBasesOpts) (blob.MerkleBase, error) {
	lgr := log.WithModule("tree-base-syncer")
	treeBaseResCh := make(chan *wire.TreeBaseRes, 1)
	iter := opts.Peers.Iterator()
	var newMerkleBase blob.MerkleBase
	for {
		peerID, ok := iter()
		if !ok {
			return newMerkleBase, ErrNoTreeBaseCandidates
		}

		var once sync.Once
		unsubTreeBaseRes := opts.Mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeTreeBaseRes, func(recvPeerID crypto.Hash, res *wire.Envelope) {
			msg := res.Message.(*wire.TreeBaseRes)
			if msg.Name != opts.Name {
				return
			}
			if peerID != recvPeerID {
				return
			}
			once.Do(func() {
				treeBaseResCh <- msg
			})
		}))
		err := opts.Mux.Send(peerID, &wire.TreeBaseReq{
			Name: opts.Name,
		})
		if err != nil {
			lgr.Warn("error fetching tree base from peer, trying another", "peer_id", peerID, "err", err)
			unsubTreeBaseRes()
			continue
		}

		timeout := 10 * time.Second
		if opts.Timeout != 0 {
			timeout = opts.Timeout
		}
		timer := time.NewTimer(timeout)

		select {
		case <-timer.C:
			lgr.Warn("timed out fetching tree base from peer, trying another", "peer_id", peerID)
			unsubTreeBaseRes()
			continue
		case msg := <-treeBaseResCh:
			unsubTreeBaseRes()
			candMerkleTree := blob.MakeTreeFromBase(msg.MerkleBase)
			if candMerkleTree.Root() != opts.MerkleRoot {
				lgr.Warn("received invalid merkle base from peer, trying another", "peer_id", peerID)
				continue
			}
			newMerkleBase = candMerkleTree.ProtocolBase()
			return newMerkleBase, nil
		}
	}
}

type SyncSectorsOpts struct {
	Timeout       time.Duration
	Mux           *p2p.PeerMuxer
	Tx            blob.Transaction
	Peers         *PeerSet
	MerkleBase    blob.MerkleBase
	SectorsNeeded []uint8
	Name          string
}

type sectorRes struct {
	peerID crypto.Hash
	msg    *wire.SectorRes
}

type reqdSectorsMap map[uint8][33]byte

func SyncSectors(opts *SyncSectorsOpts) error {
	l := log.WithModule("sector-syncer").Sub("name", opts.Name)
	tx := opts.Tx
	reqdSectors := make(reqdSectorsMap)
	for _, id := range opts.SectorsNeeded {
		hash := opts.MerkleBase[id]
		if hash == blob.EmptyBlobBaseHash {
			if err := tx.WriteSector(id, blob.ZeroSector); err != nil {
				return errors.Wrap(err, "error writing zero sector")
			}
			continue
		}
		reqdSectors[id] = awaitingSectorHash(id, hash)
	}

	neededLen := len(reqdSectors)
	var attempts int
	for {
		if attempts == 3 {
			return ErrSyncerMaxAttempts
		}

		l.Trace("performing sync attempt", "attempts", attempts+1)
		reqdSectors = syncLoop(opts, reqdSectors)
		remainingLen := len(reqdSectors)
		l.Info(
			"synced sectors",
			"received", neededLen-remainingLen,
			"remaining", remainingLen,
		)
		if remainingLen == 0 {
			return nil
		}
		if neededLen == remainingLen {
			return ErrSyncerNoProgress
		}
		neededLen = remainingLen
		attempts++
	}
}

func syncLoop(opts *SyncSectorsOpts, reqdSectors reqdSectorsMap) reqdSectorsMap {
	lgr := log.WithModule("sync-loop").Sub("name", opts.Name)

	outReqdSectors := make(map[uint8][33]byte)
	for k, v := range reqdSectors {
		outReqdSectors[k] = v
	}
	sectorReqCh := make(chan uint8)
	sectorResCh := make(chan *sectorRes)
	sectorProcessedCh := make(chan struct{}, 1)
	doneCh := make(chan struct{})
	unsubRes := opts.Mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeSectorRes, func(peerID crypto.Hash, envelope *wire.Envelope) {
		sectorResCh <- &sectorRes{
			peerID: peerID,
			msg:    envelope.Message.(*wire.SectorRes),
		}
	}))

	go func() {
		receivedSectors := make(map[uint8]bool)
		awaitingSectorID := -1
		for {
			select {
			case id := <-sectorReqCh:
				awaitingSectorID = int(id)
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
					err := opts.Mux.Send(peerID, &wire.SectorReq{
						Name:     opts.Name,
						SectorID: id,
					})
					if err != nil {
						lgr.Warn("error fetching sector from peer, trying another", "peer_id", peerID, "err", err)
						continue
					}
					lgr.Debug(
						"requested sector from peer",
						"id", id,
						"peer_id", peerID,
					)
					sendCount++
				}
			case res := <-sectorResCh:
				msg := res.msg
				peerID := res.peerID
				expHash, ok := reqdSectors[msg.SectorID]
				if msg.Name != opts.Name {
					lgr.Trace("received sector for extraneous name", "other_name", msg.Name, "sector_id", msg.SectorID)
					continue
				}
				if !ok {
					lgr.Trace("received unnecessary sector", "sector_id", msg.SectorID, "peer_id", peerID)
					continue
				}
				if receivedSectors[msg.SectorID] {
					lgr.Trace("already processed this sector", "sector_id", msg.SectorID, "peer_id", peerID)
					continue
				}
				if awaitingSectorID != int(msg.SectorID) {
					lgr.Trace("received unsolicited sector", "sector_id", msg.SectorID, "peer_id", peerID)
					continue
				}
				hash := awaitingSectorHash(msg.SectorID, blob.HashSector(msg.Sector))
				if expHash != hash {
					lgr.Warn("invalid sector received", "sector_id", msg.SectorID, "peer_id", peerID)
					continue
				}
				if err := opts.Tx.WriteSector(msg.SectorID, msg.Sector); err != nil {
					lgr.Error("failed to write sector", "sector_id", msg.SectorID, "err", err)
					continue
				}
				receivedSectors[msg.SectorID] = true
				lgr.Debug(
					"synced sector",
					"name", opts.Name,
					"sector_id", msg.SectorID,
					"peer_id", peerID,
				)
				awaitingSectorID = -1
				sectorProcessedCh <- struct{}{}
			case <-doneCh:
				return
			}
		}
	}()

sectorLoop:
	for id := range reqdSectors {
		timeout := time.NewTimer(opts.Timeout)
		lgr.Debug("requesting sector", "id", id)
		sectorReqCh <- id
		select {
		case <-sectorProcessedCh:
			lgr.Debug("sector processed", "id", id)
			delete(outReqdSectors, id)
		case <-timeout.C:
			lgr.Warn("sector request timed out", "id", id)
			break sectorLoop
		}
	}

	unsubRes()
	close(doneCh)
	return outReqdSectors
}

func awaitingSectorHash(id uint8, hash crypto.Hash) [33]byte {
	var buf [33]byte
	buf[0] = id
	copy(buf[1:], hash[:])
	return buf
}
