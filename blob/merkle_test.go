package blob

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"github.com/ddrp-org/ddrp/testutil/testfs"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

const expProof = "532a12f09febf8521419959973ad5346944c2b22bf764d0e1a34255b6564fe4b" +
	"ffe2cf7ecd1b993235746be21e91c8e61a1e22dace9850912585416501e98447" +
	"60522e01341fe96254668ba1bc2c79dd6fc66b3784c2eb39d4f07319c6232657" +
	"ae0182ab7803fc44d085c1c734a552fffdb0f744179f0d95bd60dd6f8f1818af" +
	"f34c7d70b652eba48e02b4717e1f0a2be9337b0751ccf7bf3644674cbe6423d5" +
	"3200b99bfcd82c64c818b1a2b26e14bf784fe9188a559b6b38a6dda4fb553147" +
	"b86be3a8b288b0ef0b7ee5a9852d118167a50c8471bfb9fb8f0c7992f452c79c" +
	"aabdcbb23bfceafe71f3834a17ec1d24bd4ef2daed68d2ccc3c2989fd092e353"

func TestEncodeDecode(t *testing.T) {
	f, done := testfs.NewTempFile(t)
	defer done()

	_, err := io.CopyN(f, rand.Reader, Size)
	require.NoError(t, err)

	blob := newFromFile("foobar", f)
	var buf bytes.Buffer
	merkleTree, err := Merkleize(NewReader(blob))
	require.NoError(t, err)
	require.NoError(t, merkleTree.Encode(&buf))
	require.Equal(t, 262113, buf.Len())

	otherTree := new(MerkleTree)
	require.NoError(t, otherTree.Decode(bytes.NewReader(buf.Bytes())))
	require.EqualValues(t, merkleTree, *otherTree)
}

func TestMerkleize(t *testing.T) {
	f, done := testfs.NewTempFile(t)
	defer done()
	_, err := io.CopyN(f, new(zeroReader), Size)
	require.NoError(t, err)

	blob := newFromFile("foobar", f)
	mt, err := Merkleize(NewReader(blob))
	require.NoError(t, err)
	require.Equal(t, "7d1e84e62d7ec9f6bc3f886675733da06daf0277ad8ffcf7c539938ca5ee9cf2", hex.EncodeToString(mt.Root().Bytes()))

	_, err = f.WriteAt([]byte{0x01, 0x99, 0x99, 0x99, 0xff, 0xad, 0xfc, 0x11, 0x99}, int64(SectorLen*2))
	require.NoError(t, err)

	mt, err = Merkleize(NewReader(blob))
	require.NoError(t, err)
	require.Equal(t, "e9aa8bdf96bd79398d64a31865d2dd4d95eee2ba54b06a4e30b85e9e838497dd", hex.EncodeToString(mt.Root().Bytes()))
}

func TestGenProof(t *testing.T) {
	f, done := testfs.NewTempFile(t)
	defer done()
	_, err := io.CopyN(f, new(zeroReader), Size)
	require.NoError(t, err)
	blob := newFromFile("foobar", f)
	mt, err := Merkleize(NewReader(blob))
	require.NoError(t, err)

	proof := MakeSectorProof(mt, 0)
	require.NoError(t, err)
	require.Equal(t,
		expProof,
		hex.EncodeToString(proof[:]),
	)
}

func TestVerProof(t *testing.T) {
	var proof MerkleProof
	proofB, err := hex.DecodeString(expProof)
	require.NoError(t, err)
	copy(proof[:], proofB)

	ok := VerifySectorProof(Sector{}, 0, EmptyBlobMerkleRoot, proof)
	require.True(t, ok)
}

func TestProtocolBase(t *testing.T) {
	f, done := testfs.NewTempFile(t)
	defer done()
	_, err := io.CopyN(f, new(zeroReader), Size)
	require.NoError(t, err)
	blob := newFromFile("foobar", f)
	mt, err := Merkleize(NewReader(blob))
	require.NoError(t, err)
	base := mt.ProtocolBase()

	for i := 0; i < len(base); i++ {
		require.Equal(t, "532a12f09febf8521419959973ad5346944c2b22bf764d0e1a34255b6564fe4b", hex.EncodeToString(base[i][:]))
	}
}
