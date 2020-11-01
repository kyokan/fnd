package dwire

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncodeTime(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, EncodeTime(&buf, time.Unix(12345, 0)))
	require.EqualValues(t, []byte{0x00, 0x00, 0x30, 0x39}, buf.Bytes())

	rw := new(NopReadWriter)
	err := EncodeTime(rw, 9001)
	require.Error(t, err)
	require.Contains(t, err.Error(), "value is not a time.Time")

	err = EncodeTime(rw, time.Time{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "negative UNIX time")
}

func TestDecodeTime(t *testing.T) {
	input := []byte{0x00, 0x00, 0x30, 0x39}
	rd := bytes.NewReader(input)
	var output time.Time
	require.NoError(t, DecodeTime(rd, &output))
	require.True(t, output.Equal(time.Unix(12345, 0)))

	rd = bytes.NewReader(input)
	err := DecodeTime(rd, output)
	require.Error(t, err)
	require.Contains(t, err.Error(), "value is not a *time.Time")
}
