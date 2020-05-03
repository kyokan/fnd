package wire

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEnvelope_Encoding(t *testing.T) {
	pubB, err := hex.DecodeString("02ea82c69155dec5a35eb4afbb3115e01e7f60544ab3c8ce20dfea32d9ee23ee52")
	require.NoError(t, err)
	pub, err := btcec.ParsePubKey(pubB, btcec.S256())
	require.NoError(t, err)

	var sig [65]byte
	sig[0] = 0x11
	sig[64] = 0x11
	var nonce [32]byte
	nonce[2] = 0x11

	env1 := &Envelope{
		Magic:       0xcafecafe,
		MessageType: MessageTypeHello,
		Signature:   sig,
		Timestamp:   time.Unix(0, 0),
		Message: &Hello{
			ProtocolVersion: 1,
			LocalNonce:      nonce,
			RemoteNonce:     nonce,
			PublicKey:       pub,
			UserAgent:       "someversion",
		},
	}

	var buf bytes.Buffer
	require.NoError(t, env1.Encode(&buf))
	b := buf.Bytes()
	require.Equal(t,
		"cafecafe00000000000071000000010000110000000000000000000000000000000000000000000000000000000000000011000000000000000000000000000000000000000000000000000000000002ea82c69155dec5a35eb4afbb3115e01e7f60544ab3c8ce20dfea32d9ee23ee520b736f6d6576657273696f6e1100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011",
		hex.EncodeToString(b),
	)
	env2 := &Envelope{}
	require.NoError(t, env2.Decode(bytes.NewReader(b)))
	require.Equal(t, env1.MessageType, env2.MessageType)
	require.Equal(t, env1.Timestamp.Unix(), env2.Timestamp.Unix())
	require.Equal(t, env1.Message, env2.Message)
	require.Equal(t, env1.Signature, env2.Signature)
}
