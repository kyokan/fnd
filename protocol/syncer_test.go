package protocol

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/ddrp-org/ddrp/blob"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/ddrp-org/ddrp/testutil/mockapp"
	"github.com/ddrp-org/ddrp/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type syncTreeBasesSetup struct {
	tp *mockapp.TestPeers
	ls *mockapp.TestStorage
	rs *mockapp.TestStorage
}

func TestSyncTreeBases(t *testing.T) {
	name := "foobar"
	tests := []struct {
		name string
		run  func(t *testing.T, setup *syncTreeBasesSetup)
	}{
		{
			"syncs tree bases with one valid peer",
			func(t *testing.T, setup *syncTreeBasesSetup) {
				ts := time.Now()

				var randSector blob.Sector
				_, err := rand.Read(randSector[:])
				require.NoError(t, err)
				sectorHash := blob.HashSector(randSector)
				buf := new(bytes.Buffer)
				buf.Write(blob.ZeroSector[:])
				buf.Write(randSector[:])
				update := mockapp.FillBlobReader(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					ts,
					ts,
					bytes.NewReader(buf.Bytes()),
				)

				merkleBase, err := SyncTreeBases(&SyncTreeBasesOpts{
					Mux: setup.tp.LocalMux,
					Peers: NewPeerSet([]crypto.Hash{
						crypto.HashPub(setup.tp.RemoteSigner.Pub()),
					}),
					MerkleRoot: update.MerkleRoot,
					Name:       name,
				})

				require.NoError(t, err)
				for i, hash := range merkleBase {
					if i == 1 {
						require.Equal(t, sectorHash, hash)
						continue
					}
					require.Equal(t, blob.EmptyBlobBaseHash, hash)
				}
			},
		},
		{
			"aborts sync if all peers return invalid merkle bases",
			func(t *testing.T, setup *syncTreeBasesSetup) {
				ts := time.Now()
				addlPeer, addlPeerDone := mockapp.ConnectAdditionalPeer(t, setup.tp.LocalSigner, setup.tp.LocalMux)
				defer addlPeerDone()
				addlStorage, addlStorageDone := mockapp.CreateStorage(t)
				defer addlStorageDone()
				addlSS := NewSectorServer(addlPeer.Mux, addlStorage.DB, addlStorage.BlobStore, util.NewMultiLocker())
				require.NoError(t, addlSS.Start())
				defer require.NoError(t, addlSS.Stop())
				// each blob will contain different data, thus yielding a
				// different merkle root
				mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					ts,
					ts,
				)
				mockapp.FillBlobRandom(
					t,
					addlStorage.DB,
					addlStorage.BlobStore,
					setup.tp.RemoteSigner,
					name,
					ts,
					ts,
				)

				merkleBase, err := SyncTreeBases(&SyncTreeBasesOpts{
					Mux: setup.tp.LocalMux,
					Peers: NewPeerSet([]crypto.Hash{
						crypto.HashPub(setup.tp.RemoteSigner.Pub()),
						crypto.HashPub(addlPeer.Signer.Pub()),
					}),
					MerkleRoot: crypto.Rand32(),
					Name:       name,
				})
				require.Equal(t, blob.ZeroMerkleBase, merkleBase)
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrNoTreeBaseCandidates))
			},
		},
		{
			"handles peers that time out",
			func(t *testing.T, setup *syncTreeBasesSetup) {
				ts := time.Now()
				addlPeer, addlPeerDone := mockapp.ConnectAdditionalPeer(t, setup.tp.LocalSigner, setup.tp.LocalMux)
				defer addlPeerDone()

				// trigger timeout by not configuring a sector server for the
				// additional peer

				mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					ts,
					ts,
				)

				merkleBase, err := SyncTreeBases(&SyncTreeBasesOpts{
					Timeout: 250 * time.Millisecond,
					Mux:     setup.tp.LocalMux,
					Peers: NewPeerSet([]crypto.Hash{
						crypto.HashPub(addlPeer.Signer.Pub()),
						crypto.HashPub(setup.tp.RemoteSigner.Pub()),
					}),
					MerkleRoot: crypto.Rand32(),
					Name:       name,
				})
				require.Equal(t, blob.ZeroMerkleBase, merkleBase)
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrNoTreeBaseCandidates))
			},
		},
		{
			"handles peer send errors",
			func(t *testing.T, setup *syncTreeBasesSetup) {
				ts := time.Now()
				addlPeer, addlPeerDone := mockapp.ConnectAdditionalPeer(t, setup.tp.LocalSigner, setup.tp.LocalMux)
				defer addlPeerDone()

				// trigger a peer send error by closing the additional peer
				// before starting tree base sync
				require.NoError(t, addlPeer.RemotePeer.Close())

				mockapp.FillBlobRandom(
					t,
					setup.rs.DB,
					setup.rs.BlobStore,
					setup.tp.RemoteSigner,
					name,
					ts,
					ts,
				)

				merkleBase, err := SyncTreeBases(&SyncTreeBasesOpts{
					Timeout: 250 * time.Millisecond,
					Mux:     setup.tp.LocalMux,
					Peers: NewPeerSet([]crypto.Hash{
						crypto.HashPub(addlPeer.Signer.Pub()),
						crypto.HashPub(setup.tp.RemoteSigner.Pub()),
					}),
					MerkleRoot: crypto.Rand32(),
					Name:       name,
				})
				require.Equal(t, blob.ZeroMerkleBase, merkleBase)
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrNoTreeBaseCandidates))
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

			tt.run(t, &syncTreeBasesSetup{
				tp: testPeers,
				ls: localStorage,
				rs: remoteStorage,
			})
		})
	}
}
