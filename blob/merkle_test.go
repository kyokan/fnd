package blob

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/ddrp-org/ddrp/testutil/testfs"
	"github.com/stretchr/testify/require"
)

const expProof = "d87e2783f806556be18f0d7d71324c9efa47d76dcb6ab20ebfb1bd11fcdd8e16159fb570f424dd23426b626a4274af3fb0673a3d94df9ec9ea5b32a23fb4888cd394b989e56276a3b452ffc4c693c292fd3939b3b88dbb5f07caf89434e4c598ccc660bb61bf612c077f45f098cbe1c2fc417349705b9dffb4b4357dce2839bdc587910d908005e8d451b4305190ceabff410d145a569f8d2de57c89dfac823b16b0b07f28b35127b90ee00b4cc5612ae2d2d9e2a3343ac9c0261ef96d812d43a63c26e267529de752e0f066a3a882364f28da1f02e2a1dcd852580ee804d6c4fe623c1fbe4bdad753ab5a0c281c90bcc87c59fd3c65c6723a87b72b501ae0ea"

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
	require.Equal(t, "f60de559740b01fe11dd241dee2729782caabd36613c469806a5449aaeef0a27", hex.EncodeToString(mt.Root().Bytes()))

	_, err = f.WriteAt([]byte{0x01, 0x99, 0x99, 0x99, 0xff, 0xad, 0xfc, 0x11, 0x99}, int64(SectorLen*2))
	require.NoError(t, err)

	mt, err = Merkleize(NewReader(blob))
	require.NoError(t, err)
	require.Equal(t, "75a87ea080e57925f81c6bf732bf72fb1f6465ee2ecbfaf7580b444fd3990cc7", hex.EncodeToString(mt.Root().Bytes()))
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
		require.Equal(t, "d87e2783f806556be18f0d7d71324c9efa47d76dcb6ab20ebfb1bd11fcdd8e16", hex.EncodeToString(base[i][:]))
	}
}
