package protocol

import (
	"context"
	"ddrp/crypto"
	"ddrp/p2p"
	"ddrp/testutil"
	"ddrp/testutil/testcrypto"
	"ddrp/wire"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestPinger_SendsPingsOnInterval(t *testing.T) {
	mux := p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t))
	peerID := fixedPeerID(t)
	clientConn, serverConn := testutil.NewTCPConn(t)
	peer := p2p.NewPeer(p2p.Outbound, serverConn)
	mux.AddPeer(peerID, peer)

	pingErr := make(chan error)
	go func() {
		pingErr <- PingPeer(context.Background(), &PingConfig{
			CheckInterval: 10 * time.Millisecond,
			PingInterval:  10 * time.Millisecond,
			Timeout:       10 * time.Second,
			PeerID:        peerID,
			Mux:           mux,
		})
	}()

	envelope := testutil.ReceiveEnvelope(t, clientConn)
	require.Equal(t, envelope.MessageType, wire.MessageTypePing)
	require.NoError(t, peer.Close())
	require.Equal(t, ErrPingPeerClosed, <-pingErr)
}

func TestPinger_DisconnectsAndClosesPeerIfNoPingsReceivedEver(t *testing.T) {
	mux := p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t))
	peerID := fixedPeerID(t)
	clientConn, serverConn := testutil.NewTCPConn(t)
	go func() {
		_, _ = io.Copy(ioutil.Discard, clientConn)
	}()
	peer := p2p.NewPeer(p2p.Outbound, serverConn)
	mux.AddPeer(peerID, peer)

	pingErr := make(chan error)
	go func() {
		pingErr <- PingPeer(context.Background(), &PingConfig{
			CheckInterval: 10 * time.Millisecond,
			PingInterval:  10 * time.Millisecond,
			Timeout:       40 * time.Millisecond,
			PeerID:        peerID,
			Mux:           mux,
		})
	}()

	require.Equal(t, ErrPingTimeout, <-pingErr)
	require.NotNil(t, peer.CloseReason())
}

func TestPinger_StopsAndClosesPeerIfNoPingsReceivedAfterTimeout(t *testing.T) {
	mux := p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t))
	peerID := fixedPeerID(t)
	clientConn, serverConn := testutil.NewTCPConn(t)
	go func() {
		_, _ = io.Copy(ioutil.Discard, clientConn)
	}()
	peer := p2p.NewPeer(p2p.Outbound, serverConn)
	mux.AddPeer(peerID, peer)

	pingErr := make(chan error)
	go func() {
		pingErr <- PingPeer(context.Background(), &PingConfig{
			CheckInterval: 10 * time.Millisecond,
			PingInterval:  10 * time.Millisecond,
			Timeout:       40 * time.Millisecond,
			PeerID:        peerID,
			Mux:           mux,
		})
	}()

	testutil.SendMessage(t, clientConn, wire.NewPing())
	require.Equal(t, ErrPingTimeout, <-pingErr)
	require.NotNil(t, peer.CloseReason())
}

func TestPinger_StopsIfPeerIsClosed(t *testing.T) {
	mux := p2p.NewPeerMuxer(testutil.TestMagic, testcrypto.FixedSigner(t))
	peerID := fixedPeerID(t)
	clientConn, serverConn := testutil.NewTCPConn(t)
	go func() {
		_, _ = io.Copy(ioutil.Discard, clientConn)
	}()
	peer := p2p.NewPeer(p2p.Outbound, serverConn)
	mux.AddPeer(peerID, peer)

	pingErr := make(chan error)
	go func() {
		pingErr <- PingPeer(context.Background(), &PingConfig{
			CheckInterval: 10 * time.Millisecond,
			PingInterval:  10 * time.Millisecond,
			Timeout:       10 * time.Second,
			PeerID:        peerID,
			Mux:           mux,
		})
	}()

	require.NoError(t, peer.Close())
	require.Equal(t, ErrPingPeerClosed, <-pingErr)
}

func fixedPeerID(t *testing.T) crypto.Hash {
	_, pub := testcrypto.FixedKey(t)
	return crypto.Blake2B256(pub.SerializeCompressed())
}
