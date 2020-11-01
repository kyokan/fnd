package protocol

import (
	"fnd/blob"
	"fnd/crypto"
	"fnd/p2p"
	"fnd/store"
	"fnd/testutil"
	"fnd/testutil/testcrypto"
	"fnd/util"
	"fnd/wire"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"testing"
	"time"
)

func TestUpdateServer(t *testing.T) {
	signer := testcrypto.FixedSigner(t)
	db, done := setupDB(t)
	defer done()
	mux := p2p.NewPeerMuxer(testutil.TestMagic, signer)
	nameLocker := util.NewMultiLocker()
	srv := NewUpdateServer(mux, db, nameLocker)
	require.NoError(t, srv.Start())
	defer srv.Stop()

	peerID := fixedPeerID(t)
	clientConn, serverConn := testutil.NewTCPConn(t)
	peer := p2p.NewPeer(p2p.Outbound, serverConn)
	require.NoError(t, mux.AddPeer(peerID, peer))

	tests := []struct {
		name   string
		req    *wire.UpdateReq
		setup  func(t *testing.T)
		verify func(t *testing.T)
	}{
		{
			"sends a nil update for locked names",
			&wire.UpdateReq{
				Name:      "locked",
				Timestamp: time.Now(),
			},
			func(t *testing.T) {
				require.True(t, nameLocker.TryLock("locked"))
			},
			func(t *testing.T) {
				requireNilUpdate(t, "locked", clientConn)
			},
		},
		{
			"sends a nil update for unknown names",
			&wire.UpdateReq{
				Name:      "unknown",
				Timestamp: time.Now(),
			},
			func(t *testing.T) {},
			func(t *testing.T) {
				requireNilUpdate(t, "unknown", clientConn)
			},
		},
		{
			"sends a nil update for update requests with future timestamps",
			&wire.UpdateReq{
				Name:      "future",
				Timestamp: time.Unix(10, 0),
			},
			func(t *testing.T) {
				require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:      "future",
						Timestamp: time.Unix(5, 0),
					}, blob.ZeroMerkleBase)
				}))
			},
			func(t *testing.T) {
				requireNilUpdate(t, "future", clientConn)
			},
		},
		{
			"sends a nil update for update requests with timestamps equal to stored",
			&wire.UpdateReq{
				Name:      "equal",
				Timestamp: time.Unix(10, 0),
			},
			func(t *testing.T) {
				require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:      "equal",
						Timestamp: time.Unix(10, 0),
					}, blob.ZeroMerkleBase)
				}))
			},
			func(t *testing.T) {
				requireNilUpdate(t, "equal", clientConn)
			},
		},
		{
			"sends an update for valid update requests with past timestamps",
			&wire.UpdateReq{
				Name:      "valid",
				Timestamp: time.Unix(5, 0),
			},
			func(t *testing.T) {
				ts := time.Unix(10, 0)
				tree := blob.MakeTreeFromBase(blob.ZeroMerkleBase)
				sig, err := blob.SignSeal(signer, "valid", ts, tree.Root(), crypto.ZeroHash)
				require.NoError(t, err)
				require.NoError(t, store.WithTx(db, func(tx *leveldb.Transaction) error {
					return store.SetHeaderTx(tx, &store.Header{
						Name:       "valid",
						Timestamp:  ts,
						MerkleRoot: tree.Root(),
						Signature:  sig,
					}, blob.ZeroMerkleBase)
				}))
			},
			func(t *testing.T) {
				header, err := store.GetHeader(db, "valid")
				require.NoError(t, err)
				envelope := testutil.ReceiveEnvelope(t, clientConn)
				require.EqualValues(t, &wire.Update{
					Name:       header.Name,
					Timestamp:  header.Timestamp,
					MerkleRoot: header.MerkleRoot,
					Signature:  header.Signature,
				}, envelope.Message)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			testutil.SendMessage(t, clientConn, tt.req)
			tt.verify(t)
		})
	}
}

func requireNilUpdate(t *testing.T, name string, r io.Reader) {
	envelope := testutil.ReceiveEnvelope(t, r)
	require.Equal(t, wire.MessageTypeNilUpdate, envelope.MessageType)
	require.Equal(t, name, envelope.Message.(*wire.NilUpdate).Name)
}
