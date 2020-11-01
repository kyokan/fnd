package wire

import (
	"fnd/testutil/testcrypto"
	"testing"
)

func TestHello_Encoding(t *testing.T) {
	_, pub := testcrypto.FixedKey(t)
	hello := &Hello{
		ProtocolVersion: 1,
		LocalNonce:      fixedHash,
		RemoteNonce:     fixedHash,
		PublicKey:       pub,
		UserAgent:       "foobar",
	}
	testMessageEncoding(t, "hello", hello, &Hello{})
}
