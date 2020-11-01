package mockapp

import (
	"fnd/crypto"
	"fnd/p2p"
	"fnd/testutil"
	"fnd/testutil/testcrypto"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestPeers struct {
	LocalPeer    p2p.Peer
	RemotePeer   p2p.Peer
	LocalSigner  crypto.Signer
	RemoteSigner crypto.Signer
	LocalMux     *p2p.PeerMuxer
	RemoteMux    *p2p.PeerMuxer
}

func ConnectTestPeers(t *testing.T) (*TestPeers, func()) {
	localPriv, localPub := testcrypto.RandKey()
	remotePriv, remotePub := testcrypto.RandKey()
	localSigner := crypto.NewSECP256k1Signer(localPriv)
	remoteSigner := crypto.NewSECP256k1Signer(remotePriv)

	clientConn, serverConn := testutil.NewTCPConn(t)
	localPeer := p2p.NewPeer(p2p.Outbound, clientConn)
	remotePeer := p2p.NewPeer(p2p.Inbound, serverConn)

	localMux := p2p.NewPeerMuxer(testutil.TestMagic, localSigner)
	require.NoError(t, localMux.AddPeer(crypto.HashPub(remotePub), localPeer))
	remoteMux := p2p.NewPeerMuxer(testutil.TestMagic, remoteSigner)
	require.NoError(t, remoteMux.AddPeer(crypto.HashPub(localPub), remotePeer))

	done := func() {
		require.NoError(t, localPeer.Close())
		require.NoError(t, remotePeer.Close())
	}
	return &TestPeers{
		LocalPeer:    localPeer,
		RemotePeer:   remotePeer,
		LocalSigner:  localSigner,
		RemoteSigner: remoteSigner,
		LocalMux:     localMux,
		RemoteMux:    remoteMux,
	}, done
}

type AdditionalPeer struct {
	LocalPeer  p2p.Peer
	RemotePeer p2p.Peer
	Signer     crypto.Signer
	Mux        *p2p.PeerMuxer
}

func ConnectAdditionalPeer(t *testing.T, localSigner crypto.Signer, localMux *p2p.PeerMuxer) (*AdditionalPeer, func()) {
	remotePriv, remotePub := testcrypto.RandKey()
	remoteSigner := crypto.NewSECP256k1Signer(remotePriv)

	clientConn, serverConn := testutil.NewTCPConn(t)
	localPeer := p2p.NewPeer(p2p.Outbound, clientConn)
	remotePeer := p2p.NewPeer(p2p.Inbound, serverConn)
	remoteMux := p2p.NewPeerMuxer(testutil.TestMagic, remoteSigner)
	require.NoError(t, localMux.AddPeer(crypto.HashPub(remotePub), localPeer))
	require.NoError(t, remoteMux.AddPeer(crypto.HashPub(localSigner.Pub()), remotePeer))

	done := func() {
		require.NoError(t, localPeer.Close())
		require.NoError(t, remotePeer.Close())
	}
	return &AdditionalPeer{
		LocalPeer:  localPeer,
		RemotePeer: remotePeer,
		Signer:     remoteSigner,
		Mux:        remoteMux,
	}, done
}
