package protocol

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/mslipper/handshake/primitives"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	ErrUpdateQueueMaxLen             = errors.New("update queue is at max length")
	ErrUpdateQueueIdenticalTimestamp = errors.New("timestamp is identical to stored")
	ErrUpdateQueueThrottled          = errors.New("update is throttled")
	ErrUpdateQueueStaleTimestamp     = errors.New("update is stale")
	ErrUpdateQueueSpltBrain          = errors.New("split brain")
	ErrInitialImportIncomplete       = errors.New("initial import incomplete")
)

type UpdateQueue struct {
	MaxLen   int32
	mux      *p2p.PeerMuxer
	db       *leveldb.DB
	entries  map[string]*UpdateQueueItem
	quitCh   chan struct{}
	queue    []string
	queueLen int32
	mu       sync.Mutex
	lgr      log.Logger
}

type UpdateQueueItem struct {
	PeerIDs       *PeerSet
	Name          string
	EpochHeight   uint16
	SectorSize    uint16
	SectorTipHash crypto.Hash
	ReservedRoot  crypto.Hash
	Signature     crypto.Signature
	Pub           *btcec.PublicKey
	Height        int
	Disposed      int32
}

func (u *UpdateQueueItem) Dispose() {
	atomic.StoreInt32(&u.Disposed, 1)
}

func NewUpdateQueue(mux *p2p.PeerMuxer, db *leveldb.DB) *UpdateQueue {
	return &UpdateQueue{
		MaxLen:  int32(config.DefaultConfig.Tuning.UpdateQueue.MaxLen),
		mux:     mux,
		db:      db,
		entries: make(map[string]*UpdateQueueItem),
		quitCh:  make(chan struct{}),
		lgr:     log.WithModule("update-queue"),
	}
}

func (u *UpdateQueue) Start() error {
	u.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeUpdate, u.onUpdate))
	timer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timer.C:
			u.reapDequeuedUpdates()
		case <-u.quitCh:
			return nil
		}
	}
}

func (u *UpdateQueue) Stop() error {
	close(u.quitCh)
	return nil
}

func (u *UpdateQueue) Enqueue(peerID crypto.Hash, update *wire.Update) error {
	// use atomic below to prevent having to lock mu
	// during expensive name validation calls when
	// we can cheaply check for the queue size.
	if atomic.LoadInt32(&u.queueLen) >= u.MaxLen {
		return ErrUpdateQueueMaxLen
	}

	initialImportComplete, err := store.GetInitialImportComplete(u.db)
	if err != nil {
		return errors.Wrap(err, "error getting initial import complete")
	}
	if !initialImportComplete {
		return ErrInitialImportIncomplete
	}

	nameInfo, err := u.validateUpdate(update.Name, update.EpochHeight, update.SectorSize, update.SectorTipHash, update.ReservedRoot, update.Signature)
	if err != nil {
		return errors.Wrap(err, "name failed validation")
	}

	// FIXME: epochHeight?
	var storedSectorSize uint16
	header, err := store.GetHeader(u.db, update.Name)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return errors.Wrap(err, "error getting name header")
	} else if err == nil {
		storedSectorSize = header.SectorSize
	}

	if storedSectorSize > update.SectorSize {
		return ErrUpdateQueueStaleTimestamp
	}
	if storedSectorSize == update.SectorSize {
		return ErrUpdateQueueIdenticalTimestamp
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	entry := u.entries[update.Name]
	if entry == nil || entry.SectorSize < update.SectorSize {
		u.entries[update.Name] = &UpdateQueueItem{
			PeerIDs:       NewPeerSet([]crypto.Hash{peerID}),
			Name:          update.Name,
			EpochHeight:   update.EpochHeight,
			SectorSize:    update.SectorSize,
			SectorTipHash: update.SectorTipHash,
			ReservedRoot:  update.ReservedRoot,
			Signature:     update.Signature,
			Pub:           nameInfo.PublicKey,
			Height:        nameInfo.ImportHeight,
		}

		if entry == nil {
			u.queue = append(u.queue, update.Name)
			atomic.AddInt32(&u.queueLen, 1)
		}
		u.lgr.Info("enqueued update", "name", update.Name, "epoch", update.EpochHeight, "sector", update.SectorSize)
		return nil
	}

	if entry.SectorSize > update.SectorSize {
		return ErrUpdateQueueStaleTimestamp
	}
	if entry.Signature != update.Signature {
		return ErrUpdateQueueSpltBrain
	}

	u.lgr.Info("enqueued update", "name", update.Name, "epoch", update.EpochHeight, "sector", update.SectorSize)
	entry.PeerIDs.Add(peerID)
	return nil
}

func (u *UpdateQueue) Dequeue() *UpdateQueueItem {
	u.mu.Lock()
	defer u.mu.Unlock()
	if len(u.queue) == 0 {
		return nil
	}

	name := u.queue[0]
	ret := u.entries[name]
	u.queue = u.queue[1:]
	atomic.AddInt32(&u.queueLen, -1)
	return ret
}

func (u *UpdateQueue) validateUpdate(name string, epochHeight, sectorSize uint16, mr crypto.Hash, rr crypto.Hash, sig crypto.Signature) (*store.NameInfo, error) {
	if err := primitives.ValidateName(name); err != nil {
		return nil, errors.Wrap(err, "update name is invalid")
	}
	banned, err := store.NameIsBanned(u.db, name)
	if err != nil {
		return nil, errors.Wrap(err, "error reading name ban state")
	}
	if banned {
		return nil, errors.New("name is banned")
	}
	info, err := store.GetNameInfo(u.db, name)
	if err != nil {
		return nil, errors.Wrap(err, "error reading name info")
	}
	h := blob.SealHash(name, epochHeight, sectorSize, mr, rr)
	if !crypto.VerifySigPub(info.PublicKey, sig, h) {
		return nil, errors.New("update signature is invalid")
	}
	return info, nil
}

func (u *UpdateQueue) onUpdate(peerID crypto.Hash, envelope *wire.Envelope) {
	update := envelope.Message.(*wire.Update)
	if err := u.Enqueue(peerID, update); err != nil {
		u.lgr.Info("update rejected", "name", update.Name, "reason", err)
	}
}

func (u *UpdateQueue) reapDequeuedUpdates() {
	u.mu.Lock()
	defer u.mu.Unlock()
	var toDelete []string
	for k, item := range u.entries {
		if atomic.LoadInt32(&item.Disposed) == 0 {
			continue
		}
		toDelete = append(toDelete, k)
	}
	for _, k := range toDelete {
		delete(u.entries, k)
	}
}
