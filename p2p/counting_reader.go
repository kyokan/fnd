package p2p

import (
	"io"
	"sync/atomic"
)

type CountingReader struct {
	r     io.Reader
	count uint64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{
		r: r,
	}
}

func (c *CountingReader) Count() uint64 {
	return atomic.LoadUint64(&c.count)
}

func (c *CountingReader) Reset() {
	atomic.StoreUint64(&c.count, 0)
}

func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	atomic.AddUint64(&c.count, uint64(n))
	return n, err
}
