package wire

import (
	"testing"
)

func TestUpdate_Encoding(t *testing.T) {
	update := &Update{
		Name:         "testname",
		Timestamp:    fixedTime,
		MerkleRoot:   fixedHash,
		ReservedRoot: fixedHash,
		Signature:    fixedSig,
	}

	testMessageEncoding(t, "update", update, &Update{})
}
