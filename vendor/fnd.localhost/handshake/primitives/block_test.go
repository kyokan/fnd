package primitives

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestBlock_Golden(t *testing.T) {
	tests := []struct {
		hash string
	}{
		{
			"000000000000003d56d278ef00c657de6b1fcb549f9e04e299f6a918c2573b94",
		},
		{
			"0000000000000424ee6c2a5d6e0da5edfc47a4a10328c1792056ee48303c3e40",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("block %s", tt.hash), func(t *testing.T) {
			expData, err := ioutil.ReadFile(fmt.Sprintf("testdata/block_%s.bin", tt.hash))
			require.NoError(t, err)
			block := new(Block)
			require.NoError(t, block.Decode(bytes.NewReader(expData)))
			actData := new(bytes.Buffer)
			require.NoError(t, block.Encode(actData))
			require.EqualValues(t, expData, actData.Bytes())
			genHash := block.Hash()
			require.Equal(t, tt.hash, hex.EncodeToString(genHash))
		})
	}
}
