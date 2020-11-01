package p2p

import (
	"bytes"
	"context"
	"errors"
	"fnd/testutil/testcrypto"
	"fnd/wire"
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"
	"time"
)

type blockingReadWriter struct {
	ch   chan struct{}
	err  error
	once sync.Once
}

func newBlockingReadWriter() *blockingReadWriter {
	return &blockingReadWriter{
		ch:  make(chan struct{}),
		err: errors.New("closed"),
	}
}

func (b *blockingReadWriter) Read(p []byte) (n int, err error) {
	<-b.ch
	return 0, b.err
}

func (b *blockingReadWriter) Write(p []byte) (n int, err error) {
	<-b.ch
	return 0, b.err
}

func (b *blockingReadWriter) Close() error {
	b.once.Do(func() {
		close(b.ch)
	})
	return nil
}

type brokenReadWriter struct {
	err error
}

func newBrokenReadWriter() *brokenReadWriter {
	return &brokenReadWriter{
		err: errors.New("broken"),
	}
}

func (b *brokenReadWriter) Read(p []byte) (n int, err error) {
	return 0, b.err
}

func (b *brokenReadWriter) Write(p []byte) (n int, err error) {
	return 0, b.err
}

func TestPeerImpl_Receive_HappyPath(t *testing.T) {
	expEnvelope, buf := generateEnvelope(t)
	conn := &DummyTCPConn{
		Reader: bytes.NewReader(buf),
	}
	peer := NewPeer(Outbound, conn)
	actEnvelope, err := peer.Receive()
	require.NoError(t, err)
	require.True(t, expEnvelope.Equals(actEnvelope))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := peer.Receive()
		require.Error(t, err)
		wg.Done()
	}()

	<-peer.CloseChan()
	wg.Wait()
	require.Equal(t, ErrPeerHangup, peer.CloseReason())
}

func TestPeerImpl_Receive_ContextTimeout(t *testing.T) {
	brw := newBlockingReadWriter()
	conn := &DummyTCPConn{
		Reader: brw,
	}
	peer := NewPeer(Outbound, conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, err := peer.ReceiveCtx(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context deadline exceeded")
	require.NoError(t, peer.Close())
	require.Equal(t, peer.CloseReason(), ErrPeerClosed)
}

func TestPeerImpl_Receive_ReadError(t *testing.T) {
	broken := newBrokenReadWriter()
	conn := &DummyTCPConn{
		Reader: broken,
	}
	peer := NewPeer(Outbound, conn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-peer.CloseChan()
		wg.Done()
	}()

	_, err := peer.Receive()
	require.Error(t, err)
	require.Equal(t, broken.err, err)
	wg.Wait()
	require.Equal(t, broken.err, peer.CloseReason())
}

func TestPeerImpl_Send_HappyPath(t *testing.T) {
	envelope, buf := generateEnvelope(t)
	conn := &DummyTCPConn{
		Writer: new(bytes.Buffer),
	}
	peer := NewPeer(Outbound, conn)
	require.NoError(t, peer.Send(envelope))
	require.EqualValues(t, buf, conn.Writer.(*bytes.Buffer).Bytes())
	require.NoError(t, peer.Close())
}

func TestPeerImpl_Send_ContextTimeout(t *testing.T) {
	envelope, _ := generateEnvelope(t)
	brw := newBlockingReadWriter()
	conn := &DummyTCPConn{
		Writer: brw,
	}
	peer := NewPeer(Outbound, conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	err := peer.SendCtx(ctx, envelope)
	require.Error(t, err)
	require.NoError(t, peer.Close())
	require.Equal(t, peer.CloseReason(), ErrPeerClosed)
}

func TestPeerImpl_Send_WriteError(t *testing.T) {
	envelope, _ := generateEnvelope(t)
	broken := newBrokenReadWriter()
	conn := &DummyTCPConn{
		Writer: broken,
	}
	peer := NewPeer(Outbound, conn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-peer.CloseChan()
		wg.Done()
	}()

	err := peer.Send(envelope)
	require.Error(t, err)
	require.Equal(t, broken.err, err)
	wg.Wait()
	require.Equal(t, broken.err, peer.CloseReason())
}

func TestPeerImpl_Close_ReturnsPeerClosedErrWhenClosed(t *testing.T) {
	envelope, _ := generateEnvelope(t)
	brw := newBlockingReadWriter()
	conn := &DummyTCPConn{
		Reader: brw,
		Writer: brw,
	}
	peer := NewPeer(Outbound, conn)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := peer.Send(envelope)
		require.Error(t, err)
		require.Equal(t, ErrPeerClosed, err)
		wg.Done()
	}()

	go func() {
		_, err := peer.Receive()
		require.Error(t, err)
		require.Equal(t, ErrPeerClosed, err)
		wg.Done()
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		require.Equal(t, ErrPeerClosed, peer.CloseReason())
		doneCh <- struct{}{}
	}()

	time.Sleep(5 * time.Millisecond)
	require.NoError(t, peer.Close())
	<-doneCh
}

func TestPeerImpl_Close_ReturnsPeerHangupErrWhenEOF(t *testing.T) {
	envelope, _ := generateEnvelope(t)
	brw := newBlockingReadWriter()
	brw.err = io.EOF
	conn := &DummyTCPConn{
		Reader: brw,
		Writer: brw,
	}
	peer := NewPeer(Outbound, conn)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err := peer.Send(envelope)
		require.Error(t, err)
		require.Equal(t, ErrPeerHangup, err)
		wg.Done()
	}()

	go func() {
		_, err := peer.Receive()
		require.Error(t, err)
		require.Equal(t, ErrPeerHangup, err)
		wg.Done()
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		require.Equal(t, ErrPeerHangup, peer.CloseReason())
		doneCh <- struct{}{}
	}()

	time.Sleep(5 * time.Millisecond)
	require.NoError(t, brw.Close())
	<-doneCh
}

func TestPeerImpl_Close_SendReceiveReturnsCloseReason(t *testing.T) {
	envelope, _ := generateEnvelope(t)
	brw := newBlockingReadWriter()
	brw.err = io.EOF
	conn := &DummyTCPConn{
		Reader: brw,
		Writer: brw,
	}
	peer := NewPeer(Outbound, conn)
	require.NoError(t, peer.Close())

	require.Equal(t, ErrPeerClosed, peer.Send(envelope))
	_, err := peer.Receive()
	require.Equal(t, ErrPeerClosed, err)
}

func generateEnvelope(t *testing.T) (*wire.Envelope, []byte) {
	envelope, err := wire.NewEnvelope(0, wire.NewPing(), testcrypto.NewRandomSigner())
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, envelope.Encode(&buf))
	return envelope, buf.Bytes()
}
