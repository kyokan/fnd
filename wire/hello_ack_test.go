package wire

import (
	"testing"
)

func TestHelloAck_Encoding(t *testing.T) {
	helloAck := &HelloAck{
		Nonce: fixedHash,
	}

	testMessageEncoding(t, "hello_ack", helloAck, &HelloAck{})
}
