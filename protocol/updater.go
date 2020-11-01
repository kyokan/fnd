package protocol

import (
	"fnd/blob"
	"fnd/config"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/util"
	"fnd/wire"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"time"
)

var (
	ErrUpdaterAlreadySynchronized = errors.New("updater already synchronized")
	ErrUpdaterMerkleRootMismatch  = errors.New("updater merkle root mismatch")
	ErrNameLocked                 = errors.New("name is locked")
	ErrInsufficientTimebank       = errors.New("insufficient timebank")

	updaterLogger = log.WithModule("updater")
)

type Updater struct {
	PollInterval time.Duration
	Workers      int
	mux          *p2p.PeerMuxer
	db           *leveldb.DB
	queue        *UpdateQueue
	nameLocker   util.MultiLocker
	bs           blob.Store
	obs          *util.Observable
	quitCh       chan struct{}
	wg           sync.WaitGroup
	lgr          log.Logger
}

func NewUpdater(mux *p2p.PeerMuxer, db *leveldb.DB, queue *UpdateQueue, nameLocker util.MultiLocker, bs blob.Store) *Updater {
	return &Updater{
		PollInterval: config.ConvertDuration(config.DefaultConfig.Tuning.Updater.PollIntervalMS, time.Millisecond),
		Workers:      config.DefaultConfig.Tuning.Updater.Workers,
		mux:          mux,
		db:           db,
		queue:        queue,
		nameLocker:   nameLocker,
		bs:           bs,
		obs:          util.NewObservable(),
		quitCh:       make(chan struct{}),
		lgr:          log.WithModule("updater"),
	}
}

func (u *Updater) Start() error {
	for i := 0; i < u.Workers; i++ {
		u.wg.Add(1)
		go u.runWorker()
	}
	u.wg.Wait()
	return nil
}

func (u *Updater) Stop() error {
	close(u.quitCh)
	u.wg.Wait()
	return nil
}

func (u *Updater) OnUpdateProcessed(hdlr func(item *UpdateQueueItem, err error)) util.Unsubscriber {
	return u.obs.On("update:processed", hdlr)
}

func (u *Updater) runWorker() {
	defer u.wg.Done()

	for {
		timer := time.NewTimer(u.PollInterval)
		select {
		case <-timer.C:
			item := u.queue.Dequeue()
			if item == nil {
				continue
			}

			cfg := &UpdateConfig{
				Mux:        u.mux,
				DB:         u.db,
				NameLocker: u.nameLocker,
				BlobStore:  u.bs,
				Item:       item,
			}
			if err := UpdateBlob(cfg); err != nil {
				u.obs.Emit("update:processed", item, err)
				u.lgr.Error("error processing update", "name", item.Name, "err", err)
				continue
			}
			u.obs.Emit("update:processed", item, nil)
			u.lgr.Info("name updated", "name", item.Name)
		case <-u.quitCh:
			return
		}
	}
}

type UpdateConfig struct {
	Mux        *p2p.PeerMuxer
	DB         *leveldb.DB
	NameLocker util.MultiLocker
	BlobStore  blob.Store
	Item       *UpdateQueueItem
}

func UpdateBlob(cfg *UpdateConfig) error {
	l := updaterLogger.Sub("name", cfg.Item.Name)
	item := cfg.Item
	defer item.Dispose()
	header, err := store.GetHeader(cfg.DB, item.Name)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		return errors.Wrap(err, "error getting header")
	}
	if header != nil && header.Timestamp.Equal(item.Timestamp) {
		return ErrUpdaterAlreadySynchronized
	}

	if !cfg.NameLocker.TryLock(item.Name) {
		return ErrNameLocked
	}
	defer cfg.NameLocker.Unlock(item.Name)

	newMerkleBase, err := SyncTreeBases(&SyncTreeBasesOpts{
		Timeout:    DefaultSyncerTreeBaseResTimeout,
		Mux:        cfg.Mux,
		Peers:      item.PeerIDs,
		MerkleRoot: item.MerkleRoot,
		Name:       item.Name,
	})
	if err != nil {
		return errors.Wrap(err, "error syncing merkle base")
	}

	bl, err := cfg.BlobStore.Open(item.Name)
	if err != nil {
		return errors.Wrap(err, "error getting blob")
	}
	defer func() {
		if err := bl.Close(); err != nil {
			updaterLogger.Error("error closing blob", "err", err)
		}
	}()

	var sectorsNeeded []uint8
	var prevUpdateTime time.Time
	var prevTimebank int
	var payableSectorCount int
	if header == nil {
		sectorsNeeded = blob.ZeroMerkleBase.DiffWith(newMerkleBase)
	} else {
		base, err := store.GetMerkleBase(cfg.DB, item.Name)
		if err != nil {
			return errors.Wrap(err, "error getting merkle base")
		}
		sectorsNeeded = base.DiffWith(newMerkleBase)
		prevUpdateTime = header.ReceivedAt
		prevTimebank = header.Timebank
	}
	for _, sectorID := range sectorsNeeded {
		if newMerkleBase[sectorID] == blob.EmptyBlobBaseHash {
			continue
		}
		payableSectorCount++
	}
	if payableSectorCount == 0 {
		l.Debug(
			"no payable sectors, truncating",
			"count", len(sectorsNeeded),
		)
		tx, err := bl.Transaction()
		if err != nil {
			return errors.Wrap(err, "error starting transaction")
		}
		for _, sectorID := range sectorsNeeded {
			if err := tx.WriteSector(sectorID, blob.ZeroSector); err != nil {
				return errors.Wrap(err, "error truncating sector")
			}
		}
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "error committing blob")
		}
		return nil
	}
	l.Debug(
		"calculated needed sectors",
		"total", len(sectorsNeeded),
		"payable", payableSectorCount,
	)

	newTimebank := CheckTimebank(&TimebankParams{
		TimebankDuration:     48 * time.Hour,
		MinUpdateInterval:    2 * time.Minute,
		FullUpdatesPerPeriod: 2,
	}, prevUpdateTime, prevTimebank, payableSectorCount)
	l.Debug(
		"calculated new timebank",
		"prev", prevTimebank,
		"new", newTimebank,
	)
	if newTimebank == -1 {
		return ErrInsufficientTimebank
	}

	tx, err := bl.Transaction()
	if err != nil {
		return errors.Wrap(err, "error starting transaction")
	}

	err = SyncSectors(&SyncSectorsOpts{
		Timeout:       DefaultSyncerSectorResTimeout,
		Mux:           cfg.Mux,
		Tx:            tx,
		Peers:         item.PeerIDs,
		MerkleBase:    newMerkleBase,
		SectorsNeeded: sectorsNeeded,
		Name:          item.Name,
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return errors.Wrap(err, "error during sync")
	}

	tree, err := blob.Merkleize(blob.NewReader(tx))
	if err != nil {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return errors.Wrap(err, "error calculating new blob merkle root")
	}
	if tree.Root() != item.MerkleRoot {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return ErrUpdaterMerkleRootMismatch
	}

	err = store.WithTx(cfg.DB, func(tx *leveldb.Transaction) error {
		return store.SetHeaderTx(tx, &store.Header{
			Name:         item.Name,
			Timestamp:    item.Timestamp,
			MerkleRoot:   item.MerkleRoot,
			Signature:    item.Signature,
			ReservedRoot: item.ReservedRoot,
			ReceivedAt:   time.Now(),
			Timebank:     newTimebank,
		}, tree.ProtocolBase())
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return errors.Wrap(err, "error storing header")
	}
	tx.Commit()

	height, err := store.GetLastNameImportHeight(cfg.DB)
	if err != nil {
		updaterLogger.Error("error getting last name import height, skipping gossip", "err", err)
		return nil
	}
	if height-item.Height < 10 {
		updaterLogger.Info("updated name is below gossip height, skipping", "name", item.Name)
		return nil
	}

	update := &wire.Update{
		Name:         item.Name,
		Timestamp:    item.Timestamp,
		MerkleRoot:   item.MerkleRoot,
		Signature:    item.Signature,
		ReservedRoot: item.ReservedRoot,
	}
	p2p.GossipAll(cfg.Mux, update)
	return nil
}
