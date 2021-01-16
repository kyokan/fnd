package testutil

import (
	"github.com/ddrp-org/ddrp/testutil/testcrypto"
	"github.com/ddrp-org/ddrp/wire"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
)

const (
	TestMagic = 1234
)

func RandFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)
	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

func NewTCPConn(t *testing.T) (net.Conn, net.Conn) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverCh := make(chan net.Conn)
	go func() {
		defer lis.Close()
		server, err := lis.Accept()
		if err != nil {
			panic(err)
		}
		serverCh <- server
	}()

	client, err := net.Dial("tcp", lis.Addr().String())
	require.NoError(t, err)
	return client, <-serverCh
}

func ReceiveEnvelope(t *testing.T, r io.Reader) *wire.Envelope {
	envelope := new(wire.Envelope)
	require.NoError(t, envelope.Decode(r))
	return envelope
}

func SendMessage(t *testing.T, w io.Writer, message wire.Message) {
	envelope, err := wire.NewEnvelope(TestMagic, message, testcrypto.FixedSigner(t))
	require.NoError(t, err)
	require.NoError(t, envelope.Encode(w))
}
