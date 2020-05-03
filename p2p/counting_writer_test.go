package p2p

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCountingWriter(t *testing.T) {
	var w bytes.Buffer
	cr := NewCountingWriter(&w)
	b := make([]byte, 16)
	_, err := cr.Write(b)
	require.NoError(t, err)
	assert.EqualValues(t, b, w.Bytes())
	assert.EqualValues(t, 16, cr.Count())
	cr.Reset()
	assert.EqualValues(t, 0, cr.Count())
}
