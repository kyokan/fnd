package wire

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestIPEncoder_Encoding(t *testing.T) {
	ip := net.ParseIP("127.0.0.1")
	var buf bytes.Buffer
	require.NoError(t, IPEncoder(ip).Encode(&buf))
	require.EqualValues(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x1}, buf.Bytes())

	var i IPEncoder
	require.NoError(t, i.Decode(bytes.NewReader(buf.Bytes())))
	require.True(t, ip.Equal(net.IP(i)))
}

func TestPublicKeyEncoder_Encoding(t *testing.T) {
	keyHex := "02ce4b1cf077d919934e02efd281569ad9da306d229e707c6f36572bfc645ee694"
	keyBytes, err := hex.DecodeString(keyHex)
	require.NoError(t, err)
	key, err := btcec.ParsePubKey(keyBytes, btcec.S256())
	require.NoError(t, err)
	r := bytes.NewReader(keyBytes)

	var p PublicKeyEncoder
	require.NoError(t, p.Decode(r))
	require.True(t, key.IsEqual(p.PublicKey))

	var buf bytes.Buffer
	require.NoError(t, p.Encode(&buf))
	require.Equal(t, keyHex, hex.EncodeToString(buf.Bytes()))
}
