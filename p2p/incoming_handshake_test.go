package p2p

import (
	"context"
	"errors"
	"fnd/crypto"
	"fnd/testutil/testcrypto"
	"fnd/wire"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHandleIncomingHandshake(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)
	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
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
		require.NoError(t, err)
		doneCh <- struct{}{}
	}()
	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleIncomingHandshake_InvalidHelloSig(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)
	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidEnvelopeSignature))
		doneCh <- struct{}{}
	}()
	go func() {
		_, err := HandleOutgoingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.outPeer,
			Signer:          testcrypto.NewRandomSigner(),
		})
		require.True(t, errors.Is(err, ErrPeerHangup))
		doneCh <- struct{}{}
	}()
	<-doneCh
	setup.Close(t)
	<-doneCh
}

func TestHandleIncomingHandshake_InvalidHelloAckNonce(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		err := WriteEnvelope(ctx, setup.outPeer, setup.outSigner, 12345, &wire.Hello{
			ProtocolVersion: 1,
			LocalNonce:      crypto.Rand32(),
			PublicKey:       setup.outSigner.Pub(),
		})
		require.NoError(t, err)
		_, err = setup.outPeer.Receive()
		require.NoError(t, err)
		require.NoError(t, WriteEnvelope(ctx, setup.outPeer, setup.outSigner, 12345, &wire.HelloAck{
			Nonce: [32]byte{},
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidNonce))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleIncomingHandshake_UnexpectedInitiationMessage(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		require.NoError(t, WriteEnvelope(ctx, setup.outPeer, setup.outSigner, 12345, &wire.HelloAck{
			Nonce: [32]byte{},
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.True(t, errors.Is(err, ErrUnexpectedMessage))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleIncomingHandshake_IncompatibleProtocol(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		require.NoError(t, WriteEnvelope(ctx, setup.outPeer, setup.outSigner, 12345, &wire.Hello{
			ProtocolVersion: 2,
			LocalNonce:      crypto.Rand32(),
			PublicKey:       setup.outSigner.Pub(),
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.True(t, errors.Is(err, ErrIncompatibleProtocol))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleIncomingHandshake_IncompatibleMagic(t *testing.T) {
	ctx := context.Background()
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{}, 2)

	go func() {
		require.NoError(t, WriteEnvelope(ctx, setup.outPeer, setup.outSigner, 0, &wire.Hello{
			ProtocolVersion: 1,
			LocalNonce:      crypto.Rand32(),
			PublicKey:       setup.outSigner.Pub(),
		}))
		doneCh <- struct{}{}
	}()

	go func() {
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.True(t, errors.Is(err, ErrInvalidEnvelopeMagic))
		doneCh <- struct{}{}
	}()

	<-doneCh
	<-doneCh
	setup.Close(t)
}

func TestHandleIncomingHandshake_ContextDeadlineExceeded(t *testing.T) {
	setup := initializeHandshakes(t)
	doneCh := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_, err := HandleIncomingHandshake(ctx, &HandshakeConfig{
			Magic:           12345,
			ProtocolVersion: 1,
			Peer:            setup.inPeer,
			Signer:          setup.inSigner,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
		doneCh <- struct{}{}
	}()
	<-doneCh
	setup.Close(t)
}
