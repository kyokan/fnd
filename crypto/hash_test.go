package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlake2B256(t *testing.T) {
	tests := []struct {
		in  []string
		out string
	}{
		{
			[]string{""},
			"0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8",
		},
		{
			[]string{"", "", "", ""},
			"0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8",
		},
		{
			[]string{"cafe"},
			"4e400278c29c37ee640391dfb9792390a8ac9adb6200ed47c725a86099a8586c",
		},
		{
			[]string{"0000000000000000000000000000000000000000000000000000000000000000"},
			"89eb0d6a8a691dae2cd15ed0369931ce0a949ecafa5c3f93f8121833646e15c3",
		},
		{
			[]string{"00000000000000000000000000000000", "00000000000000000000000000000000"},
			"89eb0d6a8a691dae2cd15ed0369931ce0a949ecafa5c3f93f8121833646e15c3",
		},
	}
	for _, tt := range tests {
		var pieces [][]byte
		for _, hexPiece := range tt.in {
			piece, err := hex.DecodeString(hexPiece)
			require.NoError(t, err)
			pieces = append(pieces, piece)
		}
		out := Blake2B256(pieces...).String()
		require.Equal(t, tt.out, out)
	}
}
