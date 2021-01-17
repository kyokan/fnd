package p2p

import (
	"context"
	"errors"
	"fmt"
	"fnd/crypto"
	"fnd/testutil"
	"fnd/testutil/testcrypto"
	"fnd/wire"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type handshakeSetup struct {
	inSigner  crypto.Signer
	outSigner crypto.Signer
	inPeer    Peer
	outPeer   Peer
}

func (h *handshakeSetup) Close(t *testing.T) {
	require.NoError(t, h.inPeer.Close())
	require.NoError(t, h.outPeer.Close())
}

func TestHandleOutgoingHandshake_InvalidHelloSig(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)
	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          testcrypto.NewRandomSigner(),
		})
		require.True(t, errors.Is(err, ErrPeerClosed))
		doneCh <- struct{}{}
	}()
	go func() {
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          setup.outSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidEnvelopeSignature))
		doneCh <- struct{}{}
	}()
	<-doneCh
	setup.Close(t)
	<-doneCh
}

func TestHandleOutgoingHandshake_InvalidHelloNonce(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		_, err := setup.inPeer.Receive()
		require.NoError(t, err)
		err = WriteEnvelope(ctx, setup.inPeer, setup.inSigner, 12345, &wire.Hello{
			ProtocolVersion: 1,
			LocalNonce:      crypto.Rand32(),
			PublicKey:       setup.inSigner.Pub(),
		})
		require.NoError(t, err)
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          setup.outSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidNonce))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleOutgoingHandshake_IncompatibleProtocol(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		_, err := setup.inPeer.Receive()
		require.NoError(t, err)
		require.NoError(t, WriteEnvelope(ctx, setup.inPeer, setup.inSigner, 12345, &wire.Hello{
			ProtocolVersion: 2,
			LocalNonce:      crypto.Rand32(),
			RemoteNonce:     [32]byte{},
			PublicKey:       setup.inSigner.Pub(),
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          setup.outSigner,
		})
		require.True(t, errors.Is(err, ErrIncompatibleProtocol))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleOutgoingHandshake_IncompatibleMagic(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		_, err := setup.inPeer.Receive()
		require.NoError(t, err)
		require.NoError(t, WriteEnvelope(ctx, setup.inPeer, setup.inSigner, 0, &wire.Hello{
			ProtocolVersion: 2,
			LocalNonce:      crypto.Rand32(),
			RemoteNonce:     [32]byte{},
			PublicKey:       setup.inSigner.Pub(),
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          setup.outSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidEnvelopeMagic))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleOutgoingHandshake_ContextDeadlineExceeded(t *testing.T) {
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          setup.outSigner,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
		doneCh <- struct{}{}
	}()
	<-doneCh
	setup.Close(t)
}

func initializeHandshakes(t *testing.T) *handshakeSetup {
	inPriv, _ := testcrypto.RandKey()
	outPriv, _ := testcrypto.RandKey()

	portA := testutil.RandFreePort(t)
	addrStr := fmt.Sprintf("127.0.0.1:%d", portA)
	lis, err := net.Listen("tcp", addrStr)
	require.NoError(t, err)

	connCh := make(chan net.Conn)
	go func() {
		defer lis.Close()
		conn, err := lis.Accept()
		require.NoError(t, err)
		connCh <- conn
	}()

	outConn, err := net.Dial("tcp", addrStr)
	require.NoError(t, err)

	inConn := <-connCh
	inSigner := crypto.NewSECP256k1Signer(inPriv)
	outSigner := crypto.NewSECP256k1Signer(outPriv)
	return &handshakeSetup{
		inSigner:  inSigner,
		outSigner: outSigner,
		inPeer:    NewPeer(Inbound, inConn),
		outPeer:   NewPeer(Outbound, outConn),
	}
}
