package protocol

import (
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/p2p"
	"github.com/ddrp-org/ddrp/store"
	"github.com/ddrp-org/ddrp/testutil/mockapp"
	"github.com/ddrp-org/ddrp/util"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

type updaterTestSetup struct {
	tp *mockapp.TestPeers
	ls *mockapp.TestStorage
	rs *mockapp.TestStorage
}

func TestUpdater(t *testing.T) {
	name := "foobar"
	tests := []struct {
		name string
		run  func(t *testing.T, setup *updaterTestSetup)
	}{
		{
			"syncs sectors when the local node has never seen the blob",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					CurrentEpoch(name),
					blob.SectorCount,
					ts,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"syncs sectors when the local node has an older blob",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				epochHeight := CurrentEpoch(name)
				sectorSize := uint16(10)
				mockapp.FillBlobReader(
					t,
					setup.ls.DB,
					setup.ls.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize,
					ts.Add(-48*time.Hour),
					mockapp.NullReader,
				)
				// create the new blob remotely
				update := mockapp.FillBlobReader(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize+10,
					ts,
					mockapp.NullReader,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"aborts sync when there is a sector tip hash mismatch",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				epochHeight := CurrentEpoch(name)
				sectorSize := uint16(10)
				mockapp.FillBlobReader(
					t,
					setup.ls.DB,
					setup.ls.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize,
					ts.Add(-48*time.Hour),
					rand.Reader,
				)
				// create the new blob remotely
				update := mockapp.FillBlobReader(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize+10,
					ts,
					rand.Reader,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
					},
				}
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrUpdaterSectorTipHashMismatch))
				header, err := store.GetHeader(setup.ls.DB, name)
				require.True(t, header.Banned)
			},
		},
		{
			"aborts sync if the new sector size is equal to the stored sector size",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				epochHeight := CurrentEpoch(name)
				sectorSize := uint16(0)
				update := mockapp.FillBlobRandom(
					t,
					setup.ls.DB,
					setup.ls.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize,
					ts,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: epochHeight,
						SectorSize:  sectorSize,
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrUpdaterAlreadySynchronized))
			},
		},
		{
			"aborts sync if the name is locked",
			func(t *testing.T, setup *updaterTestSetup) {
				locker := util.NewMultiLocker()
				require.True(t, locker.TryLock(name))
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: locker,
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:        name,
						EpochHeight: CurrentEpoch(name),
					},
				}
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrNameLocked))
			},
		},
		{
			"does not gossip if the name has fewer than 10 confirmations",
			func(t *testing.T, setup *updaterTestSetup) {
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetLastNameImportHeightTx(tx, 100)
				}))
				ts := time.Now()
				epochHeight := CurrentEpoch(name)
				sectorSize := uint16(0)
				updateCh := make(chan struct{})
				unsub := setup.tp.RemoteMux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeUpdate, func(id crypto.Hash, envelope *wire.Envelope) {
					updateCh <- struct{}{}
				}))
				defer unsub()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize,
					ts,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
						Height:        101,
					},
				}
				require.NoError(t, UpdateBlob(cfg))

				timeout := time.NewTimer(100 * time.Millisecond)
				select {
				case <-updateCh:
					t.FailNow()
				case <-timeout.C:
				}
			},
		},
		{
			"gossips if the name has more than 10 confirmations",
			func(t *testing.T, setup *updaterTestSetup) {
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetLastNameImportHeightTx(tx, 100)
				}))
				ts := time.Now()
				epochHeight := CurrentEpoch(name)
				sectorSize := uint16(0)
				updateCh := make(chan *wire.Envelope, 1)
				unsub := setup.tp.RemoteMux.AddMessageHandler(p2p.PeerMessageHandlerForType(wire.MessageTypeUpdate, func(id crypto.Hash, envelope *wire.Envelope) {
					updateCh <- envelope
				}))
				defer unsub()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					epochHeight,
					sectorSize,
					ts,
				)
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:          name,
						EpochHeight:   update.EpochHeight,
						SectorSize:    update.SectorSize,
						SectorTipHash: update.SectorTipHash,
						ReservedRoot:  update.ReservedRoot,
						Signature:     update.Signature,
						Pub:           setup.tp.RemoteSigner.Pub(),
						Height:        80,
					},
				}
				require.NoError(t, UpdateBlob(cfg))

				timeout := time.NewTimer(250 * time.Millisecond)
				select {
				case envelope := <-updateCh:
					msg := envelope.Message.(*wire.Update)
					require.Equal(t, name, msg.Name)
				case <-timeout.C:
					t.Fail()
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPeers, peersDone := mockapp.ConnectTestPeers(t)
			defer peersDone()
			remoteStorage, remoteStorageDone := mockapp.CreateStorage(t)
			defer remoteStorageDone()
			localStorage, localStorageDone := mockapp.CreateStorage(t)
			defer localStorageDone()
			remoteSS := NewSectorServer(testPeers.RemoteMux, remoteStorage.DB, remoteStorage.BlobStore, util.NewMultiLocker())
			require.NoError(t, remoteSS.Start())
			defer require.NoError(t, remoteSS.Stop())

			tt.run(t, &updaterTestSetup{
				tp: testPeers,
				ls: localStorage,
				rs: remoteStorage,
			})
		})
	}
}
