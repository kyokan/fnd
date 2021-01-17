package wire

import (
	"bytes"
	"sync"

	"fnd/crypto"
	"fnd/dwire"
)

var hashCacherBufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type HashCacher struct {
	hash crypto.Hash
	once sync.Once
}

func (h *HashCacher) Hash(enc dwire.Encoder) (crypto.Hash, error) {
	h.once.Do(func() {
		buf := hashCacherBufPool.Get().(*bytes.Buffer)
		buf.Reset()
		if err := enc.Encode(buf); err != nil {
			panic(err)
		}
		h.hash = crypto.Blake2B256(buf.Bytes())
		hashCacherBufPool.Put(buf)
	})
	return h.hash, nil
}
