package protocol

import (
	"crypto/rand"
	"errors"
	"fnd/blob"
	"fnd/crypto"
	"fnd/store"
	"fnd/testutil/mockapp"
	"fnd/util"
	"testing"
	"time"

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
					blob.SectorBytes,
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"aborts sync when there is an equivocation",
			func(t *testing.T, setup *updaterTestSetup) {
				require.NoError(t, store.WithTx(setup.rs.DB, func(tx *leveldb.Transaction) error {
					if err := store.SetInitialImportCompleteTx(tx); err != nil {
						return err
					}
					if err := store.SetNameInfoTx(tx, name, setup.tp.RemoteSigner.Pub(), 10); err != nil {
						return err
					}
					return nil
				}))
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
				// this forces an equivocation because local has 10 random sectors
				// and remote as 20 _different_ random sectors. Prev Hash at 10 will
				// mismatch, leading to ErrInvalidPrevHash
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				time.Sleep(time.Second)
				require.True(t, errors.Is(err, ErrPayloadEquivocation))
			},
		},
		{
			"aborts sync when there is a invalid payload signature",
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
				// The sectors are generated from null reader, so there won't
				// be an equivocation, however we use local signer on the
				// remote (instead of remote signer like other test cases
				// above), so it generates an invalid signature.
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					// TODO: ideally setup a fake signer
					if err := store.SetNameInfoTx(tx, name, setup.tp.LocalSigner.Pub(), 10); err != nil {
						return err
					}
					return nil
				}))
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidPayloadSignature))
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
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
			localSS := NewSectorServer(testPeers.LocalMux, localStorage.DB, localStorage.BlobStore, util.NewMultiLocker())
			require.NoError(t, localSS.Start())
			defer require.NoError(t, localSS.Stop())
			go func() {
				remoteUQ := NewUpdateQueue(testPeers.RemoteMux, localStorage.DB)
				require.NoError(t, remoteUQ.Start())
				defer require.NoError(t, remoteUQ.Stop())
			}()

			require.NoError(t, store.WithTx(localStorage.DB, func(tx *leveldb.Transaction) error {
				if err := store.SetInitialImportCompleteTx(tx); err != nil {
					return err
				}
				if err := store.SetNameInfoTx(tx, name, testPeers.RemoteSigner.Pub(), 10); err != nil {
					return err
				}
				return nil
			}))

			tt.run(t, &updaterTestSetup{
				tp: testPeers,
				ls: localStorage,
				rs: remoteStorage,
			})
		})
	}
}

func TestEpoch(t *testing.T) {
	name := "foobar"
	tests := []struct {
		name string
		run  func(t *testing.T, setup *updaterTestSetup)
	}{
		{
			"syncs sectors when the local node has never seen the name before",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					0,
					blob.SectorBytes,
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"aborts sync if the name is banned",
			func(t *testing.T, setup *updaterTestSetup) {
				cfg := &UpdateConfig{
					Mux:       setup.tp.LocalMux,
					DB:        setup.ls.DB,
					BlobStore: setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:        name,
						EpochHeight: CurrentEpoch(name) + 1,
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					err := store.SetHeaderBan(tx, name, time.Time{})
					if err != nil {
						return err
					}
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: 0,
						SectorSize:  10,
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrNameBanned))
			},
		},
		{
			"syncs sectors when the name ban has passed",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					CurrentEpoch(name)+1,
					blob.SectorBytes,
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					err := store.SetHeaderBan(tx, name, time.Now().Add(-10*24*time.Duration(time.Hour)))
					if err != nil {
						return err
					}
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: 0,
						SectorSize:  0,
					}, blob.ZeroSectorHashes)
				}))
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"aborts sync if the epoch is throttled",
			func(t *testing.T, setup *updaterTestSetup) {
				cfg := &UpdateConfig{
					Mux:        setup.tp.LocalMux,
					DB:         setup.ls.DB,
					NameLocker: util.NewMultiLocker(),
					BlobStore:  setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:        name,
						EpochHeight: CurrentEpoch(name) - 1,
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:         name,
						EpochHeight:  0,
						SectorSize:   10,
						EpochStartAt: time.Now(),
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidEpochThrottled))
			},
		},
		{
			"aborts sync if the epoch is backdated",
			func(t *testing.T, setup *updaterTestSetup) {
				cfg := &UpdateConfig{
					Mux:       setup.tp.LocalMux,
					DB:        setup.ls.DB,
					BlobStore: setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:        name,
						EpochHeight: 0,
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: CurrentEpoch(name),
						SectorSize:  10,
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidEpochBackdated))
			},
		},
		{
			"aborts sync if the epoch is futuredated",
			func(t *testing.T, setup *updaterTestSetup) {
				cfg := &UpdateConfig{
					Mux:       setup.tp.LocalMux,
					DB:        setup.ls.DB,
					BlobStore: setup.ls.BlobStore,
					Item: &UpdateQueueItem{
						PeerIDs: NewPeerSet([]crypto.Hash{
							crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						}),
						Name:        name,
						EpochHeight: CurrentEpoch(name) + 2,
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: CurrentEpoch(name),
						SectorSize:  10,
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidEpochFuturedated))
			},
		},
		{
			"rewrites partial blob with new blob on epoch rollover",
			func(t *testing.T, setup *updaterTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					CurrentEpoch(name),
					blob.SectorBytes,
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
						Name:        name,
						EpochHeight: update.EpochHeight,
						SectorSize:  update.SectorSize,
						Pub:         setup.tp.RemoteSigner.Pub(),
					},
				}
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: CurrentEpoch(name) - 1,
						SectorSize:  10,
					}, blob.ZeroSectorHashes)
				}))
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
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

			require.NoError(t, store.WithTx(localStorage.DB, func(tx *leveldb.Transaction) error {
				if err := store.SetInitialImportCompleteTx(tx); err != nil {
					return err
				}
				if err := store.SetNameInfoTx(tx, name, testPeers.RemoteSigner.Pub(), 10); err != nil {
					return err
				}
				return nil
			}))

			tt.run(t, &updaterTestSetup{
				tp: testPeers,
				ls: localStorage,
				rs: remoteStorage,
			})
		})
	}
}
