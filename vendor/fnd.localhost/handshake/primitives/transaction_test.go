package primitives

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestTransaction_Golden(t *testing.T) {
	tests := []struct {
		txid string
	}{
		{
			"1d0f8de2757488cbd59bea7b8f7c7ad5aa9ebd6459631e801a041062338a8630",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("transaction %s", tt.txid), func(t *testing.T) {
			expData, err := ioutil.ReadFile(fmt.Sprintf("testdata/tx_%s.bin", tt.txid))
			require.NoError(t, err)
			tx := new(Transaction)
			require.NoError(t, tx.Decode(bytes.NewReader(expData)))
			actData := new(bytes.Buffer)
			require.NoError(t, tx.Encode(actData))
			require.EqualValues(t, expData, actData.Bytes())

			id := tx.ID()
			require.Equal(t, tt.txid, hex.EncodeToString(id))
		})
	}
}
