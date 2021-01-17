package protocol

import (
	"errors"
	"testing"
	"time"

	"fnd/blob"
	"fnd/crypto"
	"fnd/store"
	"fnd/testutil/mockapp"
	"fnd/util"

	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
)

type epochTestSetup struct {
	tp *mockapp.TestPeers
	ls *mockapp.TestStorage
	rs *mockapp.TestStorage
}

func TestEpoch(t *testing.T) {
	name := "foobar"
	tests := []struct {
		name string
		run  func(t *testing.T, setup *epochTestSetup)
	}{
		{
			"syncs sectors when the local node has never seen the name before",
			func(t *testing.T, setup *epochTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					0,
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
			"aborts sync if the name is banned",
			func(t *testing.T, setup *epochTestSetup) {
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
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: 0,
						SectorSize:  10,
						Banned:      true,
						BannedAt:    time.Now().Add(-1 * 24 * time.Duration(time.Hour)),
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrNameBanned))
			},
		},
		{
			"syncs sectors when the name ban has passed",
			func(t *testing.T, setup *epochTestSetup) {
				ts := time.Now()
				update := mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					0,
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
				require.NoError(t, store.WithTx(setup.ls.DB, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:        name,
						EpochHeight: 0,
						SectorSize:  0,
						Banned:      true,
						BannedAt:    time.Now().Add(-8 * 24 * time.Duration(time.Hour)),
					}, blob.ZeroSectorHashes)
				}))
				require.NoError(t, UpdateBlob(cfg))
				mockapp.RequireBlobsEqual(t, setup.ls.BlobStore, setup.rs.BlobStore, name)
			},
		},
		{
			"aborts sync if the epoch is throttled",
			func(t *testing.T, setup *epochTestSetup) {
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
			func(t *testing.T, setup *epochTestSetup) {
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
						Banned:      true,
						BannedAt:    time.Now().Add(-1 * 24 * time.Duration(time.Hour)),
					}, blob.ZeroSectorHashes)
				}))
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidEpochBackdated))
			},
		},
		{
			"aborts sync if the epoch is futuredated",
			func(t *testing.T, setup *epochTestSetup) {
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
				err := UpdateBlob(cfg)
				require.NotNil(t, err)
				require.True(t, errors.Is(err, ErrInvalidEpochFuturedated))
			},
		},
		{
			"rewrites partial blob with new blob on epoch rollover",
			func(t *testing.T, setup *epochTestSetup) {
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

			tt.run(t, &epochTestSetup{
				tp: testPeers,
				ls: localStorage,
				rs: remoteStorage,
			})
		})
	}
}
