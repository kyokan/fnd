package wire

import (
	"encoding/hex"
	"io"
	"sync"
	"testing"

	"fnd/crypto"
	"fnd/dwire"

	"github.com/stretchr/testify/require"
)

type input struct {
	HashCacher

	data string
}

func (i *input) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		i.data,
	)
}

func (i *input) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&i.data,
	)
}

func (i *input) Hash() (crypto.Hash, error) {
	return i.HashCacher.Hash(i)
}

func TestHashCacher_Hash(t *testing.T) {
	in := &input{
		data: "hello",
	}

	runs := 10
	var wg sync.WaitGroup
	wg.Add(runs)
	for i := 0; i < runs; i++ {
		go func() {
			h, err := in.Hash()
			require.NoError(t, err)
			require.Equal(t, "1cac82bdb18fa434f7af6bd97d5ee7dbd17e45c8eb8921a874427a0886e5a93a", hex.EncodeToString(h[:]))
			wg.Done()
		}()
	}

	wg.Wait()
}
