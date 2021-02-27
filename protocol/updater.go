package protocol

import (
	"fnd/blob"
	"fnd/config"
	"fnd/crypto"
	"fnd/log"
	"fnd/p2p"
	"fnd/store"
	"fnd/util"
	"fnd/wire"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	ErrUpdaterAlreadySynchronized   = errors.New("updater already synchronized")
	ErrUpdaterSectorTipHashMismatch = errors.New("updater sector tip hash mismatch")
	ErrNameLocked                   = errors.New("name is locked")
	ErrNameBanned                   = errors.New("name is banned")
	ErrInvalidEpochCurrent          = errors.New("name epoch invalid current")
	ErrInvalidEpochThrottled        = errors.New("name epoch invalid throttled")
	ErrInvalidEpochBackdated        = errors.New("name epoch invalid backdated")
	ErrInvalidEpochFuturedated      = errors.New("name epoch invalid futuredated")

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

	// If the new update and is the same size or fewer reject if it is in the
	// same epoch. In the future, this may be an equivocation condition
	// may not be eq
	if header != nil && header.EpochHeight == item.EpochHeight && header.SectorSize >= item.SectorSize {
		return ErrUpdaterAlreadySynchronized
	}

	var prevHash crypto.Hash = blob.ZeroHash
	var epochHeight, sectorSize uint16
	var epochUpdated bool
	if header != nil {
		epochHeight = header.EpochHeight
		sectorSize = header.SectorSize
		prevHash = header.SectorTipHash
	}

	// header is the existing header/data in the db
	// item is the new incoming update

	// The new header should have a higher or equal epoch
	if item.EpochHeight < epochHeight {
		return ErrInvalidEpochBackdated
	}

	bannedAt, err := store.GetHeaderBan(cfg.DB, item.Name)
	if err != nil {
		return err
	}

	// If it is higher (skip if it's appending data to the same epoch)
	if header != nil && item.EpochHeight > epochHeight {
		// Recovery from banning must increment the epoch by at least 2 and one
		// real week since the local node banned
		if !bannedAt.IsZero() {
			// Banned for at least a week
			if bannedAt.Add(7 * 24 * time.Duration(time.Hour)).After(time.Now()) {
				return ErrNameBanned
			}

			// Publisher is banned for the equivocating epoch and the next epoch
			// The faulty epoch may be old or backdated, so the penalty may not be
			// as large as it seems
			if item.EpochHeight <= epochHeight+1 {
				return ErrInvalidEpochCurrent
			}
		}

		// If the epoch is updated less than a week ago BUT NOT the current
		// epoch or the next one. The node can bank up one extra epoch just in
		// case (or periodically burst and do two epochs in a week). This
		// conditions is only valid if the last local epoch increment is less
		// than a week old.
		if time.Now().Before(header.EpochStartAt.Add(7 * 24 * time.Duration(time.Hour))) {
			if item.EpochHeight < CurrentEpoch(item.Name)+1 {
				return ErrInvalidEpochThrottled
			}
		}

		// Reject any epochs more than one in the future
		if item.EpochHeight > CurrentEpoch(item.Name)+1 {
			return ErrInvalidEpochFuturedated
		}

		// Sync the entire blob on epoch rollover
		epochUpdated = true
		sectorSize = 0
	}

	// check blob res prev hash and equivocate

	if !cfg.NameLocker.TryLock(item.Name) {
		return ErrNameLocked
	}
	defer cfg.NameLocker.Unlock(item.Name)

	bl, err := cfg.BlobStore.Open(item.Name)
	if err != nil {
		return errors.Wrap(err, "error getting blob")
	}
	defer func() {
		if err := bl.Close(); err != nil {
			updaterLogger.Error("error closing blob", "err", err)
		}
	}()

	tx, err := bl.Transaction()
	if err != nil {
		return errors.Wrap(err, "error starting transaction")
	}

	_, err = tx.Seek(int64(sectorSize)*int64(blob.SectorBytes), io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "error seeking transaction")
	}

	err = SyncSectors(&SyncSectorsOpts{
		Timeout:     DefaultSyncerBlobResTimeout,
		Mux:         cfg.Mux,
		Tx:          tx,
		Peers:       item.PeerIDs,
		EpochHeight: epochHeight,
		SectorSize:  sectorSize,
		PrevHash:    prevHash,
		Name:        item.Name,
		DB:          cfg.DB,
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return errors.Wrap(err, "error during sync")
	}
	tree, err := blob.SerialHash(blob.NewReader(tx), blob.ZeroHash, item.SectorSize)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			updaterLogger.Error("error rolling back blob transaction", "err", err)
		}
		return errors.Wrap(err, "error calculating new blob sector tip hash")
	}

	var sectorsNeeded uint16

	if header == nil {
		sectorsNeeded = item.SectorSize
	} else {
		sectorsNeeded = item.SectorSize - header.SectorSize
	}
	l.Debug(
		"calculated needed sectors",
		"total", sectorsNeeded,
	)

	var epochStart time.Time
	if epochUpdated {
		epochStart = time.Now()
	}

	err = store.WithTx(cfg.DB, func(tx *leveldb.Transaction) error {
		return store.SetHeaderTx(tx, &store.Header{
			Name:          item.Name,
			EpochHeight:   item.EpochHeight,
			SectorSize:    item.SectorSize,
			SectorTipHash: tree.Tip(),
			EpochStartAt:  epochStart,
		}, tree)
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
		Name:        item.Name,
		EpochHeight: item.EpochHeight,
		SectorSize:  item.SectorSize,
	}
	p2p.GossipAll(cfg.Mux, update)
	return nil
}
