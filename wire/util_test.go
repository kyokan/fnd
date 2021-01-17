package wire

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"fnd/blob"
	"fnd/crypto"

	"github.com/stretchr/testify/require"
)

var (
	fixedSig         crypto.Signature
	fixedHash        crypto.Hash
	fixedMerkleProof blob.MerkleProof
	fixedEpochHeight = uint16(0)
	fixedSectorSize  = uint16(0)
)

func testMessageEncoding(t *testing.T, fixtureName string, input Message, proto interface{}) {
	fixtureData, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", fixtureName))
	require.NoError(t, err)
	protoMsg, ok := proto.(Message)
	require.True(t, ok)
	require.NoError(t, protoMsg.Decode(bytes.NewReader(fixtureData)))
	require.True(t, input.Equals(protoMsg))

	var buf bytes.Buffer
	require.NoError(t, input.Encode(&buf))

	if len(fixtureData) == 0 {
		require.Nil(t, buf.Bytes())
	} else {
		require.EqualValues(t, fixtureData, buf.Bytes())
	}

	expHash := crypto.Blake2B256(fixtureData)
	inputHash, err := input.Hash()
	require.NoError(t, err)
	protoHash, err := protoMsg.Hash()
	require.NoError(t, err)
	require.Equal(t, expHash, inputHash)
	require.Equal(t, expHash, protoHash)
}

func init() {
	fixedSig[1] = 0xff
	fixedHash[1] = 0xff
	fixedMerkleProof[1] = 0xff
}
