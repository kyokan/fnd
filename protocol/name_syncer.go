package protocol

import (
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/log"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/util"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNameSyncerSyncResponseTimeout = errors.New("sync timed out")
)

type NameSyncer struct {
	Workers               int
	SampleSize            int
	UpdateResponseTimeout time.Duration
	Interval              time.Duration
	SyncResponseTimeout   time.Duration
	mux                   *p2p.PeerMuxer
	db                    *leveldb.DB
	nameLocker            util.MultiLocker
	updater               *Updater
	obs                   *util.Observable
	lgr                   log.Logger
	doneCh                chan struct{}
	once                  sync.Once
}

func NewNameSyncer(mux *p2p.PeerMuxer, db *leveldb.DB, nameLocker util.MultiLocker, updater *Updater) *NameSyncer {
	return &NameSyncer{
		Workers:               config.DefaultConfig.Tuning.NameSyncer.Workers,
		SampleSize:            config.DefaultConfig.Tuning.NameSyncer.SampleSize,
		UpdateResponseTimeout: config.ConvertDuration(config.DefaultConfig.Tuning.NameSyncer.UpdateResponseTimeoutMS, time.Millisecond),
		Interval:              config.ConvertDuration(config.DefaultConfig.Tuning.NameSyncer.IntervalMS, time.Millisecond),
		SyncResponseTimeout:   config.ConvertDuration(config.DefaultConfig.Tuning.NameSyncer.SyncResponseTimeoutMS, time.Millisecond),
		mux:                   mux,
		db:                    db,
		nameLocker:            nameLocker,
		updater:               updater,
		obs:                   util.NewObservable(),
		lgr:                   log.WithModule("name-syncer"),
		doneCh:                make(chan struct{}),
	}
}

func (ns *NameSyncer) Start() error {
	ns.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeUpdate, ns.handleUpdate))
	ns.mux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeNilUpdate, ns.handleNilUpdate))

	resyncTick := time.NewTicker(ns.Interval)
	for {
		in, out := ns.mux.PeerCount()
		total := in + out
		if total == 0 {
			ns.lgr.Info("no connected peers, skipping name sync")
			time.Sleep(5 * time.Second)
			continue
		}
		initialImportComplete, err := store.GetInitialImportComplete(ns.db)
		if err != nil {
			ns.lgr.Error("error getting initial import complete", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if !initialImportComplete {
			ns.lgr.Info("initial import incomplete, skipping name sync")
			time.Sleep(time.Minute)
			continue
		}
		ns.doSync()
		break
	}

	for {
		select {
		case <-resyncTick.C:
			ns.doSync()
		case <-ns.doneCh:
			return nil
		}
	}
}

func (ns *NameSyncer) Stop() error {
	close(ns.doneCh)
	return nil
}

func (ns *NameSyncer) handleUpdate(peerID crypto.Hash, envelope *wire.Envelope) {
	ns.obs.Emit("message:update", envelope.Message.(*wire.Update))
}

func (ns *NameSyncer) handleNilUpdate(peerID crypto.Hash, envelope *wire.Envelope) {
	ns.obs.Emit("message:nil-update", envelope.Message.(*wire.NilUpdate))
}

func (ns *NameSyncer) onUpdate(name string, hdlr func()) util.Unsubscriber {
	return ns.obs.On("message:update", func(update *wire.Update) {
		if update.Name != name {
			return
		}
		hdlr()
	})
}

func (ns *NameSyncer) onNilUpdate(name string, hdlr func()) util.Unsubscriber {
	return ns.obs.On("message:nil-update", func(update *wire.NilUpdate) {
		if update.Name != name {
			return
		}
		hdlr()
	})
}

func (ns *NameSyncer) doSync() {
	ns.lgr.Info("starting name sync")

	var syncCount int64
	sem := make(chan struct{}, ns.Workers)

	stream, err := store.StreamNameInfo(ns.db, "")
	if err != nil {
		ns.lgr.Error("error opening name info stream", "err", err)
		return
	}
	defer stream.Close()

	for {
		info, err := stream.Next()
		if err != nil {
			ns.lgr.Error("error reading name info", "err", err)
			return
		}
		if info == nil {
			break
		}

		sem <- struct{}{}
		go func() {
			ns.syncName(info)
			atomic.AddInt64(&syncCount, 1)
			<-sem
		}()
	}
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	ns.obs.Emit("sync:job-complete", int(syncCount))
	ns.lgr.Info("finished name sync", "count", syncCount)
}

func (ns *NameSyncer) OnNameComplete(cb func(name string, receiptCount int)) util.Unsubscriber {
	return ns.obs.On("sync:name-complete", cb)
}

func (ns *NameSyncer) OnJobComplete(cb func(count int)) util.Unsubscriber {
	return ns.obs.On("sync:job-complete", cb)
}

func (ns *NameSyncer) OnSyncError(cb func(name string, err error)) util.Unsubscriber {
	return ns.obs.On("sync:err", cb)
}

func (ns *NameSyncer) syncName(info *store.NameInfo) {
	name := info.Name
	ownTS := time.Unix(0, 0)
	header, err := store.GetHeader(ns.db, info.Name)
	if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
		ns.lgr.Error(
			"failed to fetch name header",
			"err", err,
		)
		return
	}
	if err == nil {
		ownTS = header.Timestamp
	}

	isEnvelopeCh := make(chan bool)
	envelopeCountCh := make(chan int)
	doneCh := make(chan struct{})
	unsubUpdate := ns.onUpdate(name, func() {
		isEnvelopeCh <- true
	})
	unsubNilUpdate := ns.onNilUpdate(name, func() {
		isEnvelopeCh <- false
	})

	recips, _ := p2p.BroadcastRandom(ns.mux, ns.SampleSize, &wire.UpdateReq{
		Name:      name,
		Timestamp: ownTS,
	})
	sampleSize := len(recips)

	go func() {
		timeout := time.NewTimer(ns.UpdateResponseTimeout)
		var recvCount int
		var envelopeCount int
		var fired bool
		for {
			select {
			case isEnvelope := <-isEnvelopeCh:
				recvCount++
				if isEnvelope {
					envelopeCount++
				}
				if !fired && recvCount == sampleSize {
					fired = true
					envelopeCountCh <- envelopeCount
				}
			case <-timeout.C:
				if !fired {
					fired = true
					envelopeCountCh <- envelopeCount
				}
			case <-doneCh:
				return
			}
		}
	}()

	envelopeCount := <-envelopeCountCh
	unsubUpdate()
	unsubNilUpdate()
	doneCh <- struct{}{}
	ns.awaitSyncCompletion(name, envelopeCount, sampleSize)
}

func (ns *NameSyncer) awaitSyncCompletion(name string, receiptCount int, sampleSize int) {
	if receiptCount == 0 {
		ns.obs.Emit("sync:name-complete", name, receiptCount)
		ns.lgr.Info(
			"synced name",
			"name", name,
			"receipt_count", receiptCount,
			"nil_count", sampleSize-receiptCount,
		)
		return
	}

	var once sync.Once
	errCh := make(chan error)
	unsub := ns.updater.OnUpdateProcessed(func(item *UpdateQueueItem, err error) {
		if item.Name != name {
			return
		}
		once.Do(func() {
			errCh <- err
		})
	})
	defer unsub()

	timeout := time.NewTimer(ns.SyncResponseTimeout)

	select {
	case <-timeout.C:
		ns.obs.Emit("sync:err", name, ErrNameSyncerSyncResponseTimeout)
		ns.lgr.Error("sync timed out", "name", name)
		return
	case err := <-errCh:
		if err != nil {
			ns.obs.Emit("sync:err", name, errors.Wrap(err, "failed to sync name"))
			ns.lgr.Error("encountered error while syncing name", "err", err)
			return
		}
		ns.obs.Emit("sync:name-complete", name, receiptCount)
		ns.lgr.Info(
			"synced name",
			"name", name,
			"receipt_count", receiptCount,
			"nil_count", sampleSize-receiptCount,
		)
	}
}
