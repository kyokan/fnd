package p2p

import (
	"bytes"
	"github.com/ddrp-org/ddrp/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCountingReader(t *testing.T) {
	buf := crypto.Rand32()
	r := bytes.NewReader(buf[:])
	cr := NewCountingReader(r)
	b := make([]byte, 16)
	_, err := cr.Read(b)
	assert.EqualValues(t, buf[:16], b)
	require.NoError(t, err)
	assert.EqualValues(t, 16, cr.Count())
	cr.Reset()
	assert.EqualValues(t, 0, cr.Count())
}
