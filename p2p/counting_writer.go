package p2p

import (
	"io"
	"sync/atomic"
)

type CountingWriter struct {
	w     io.Writer
	count uint64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{
		w: w,
	}
}

func (c *CountingWriter) Count() uint64 {
	return atomic.LoadUint64(&c.count)
}

func (c *CountingWriter) Reset() {
	atomic.StoreUint64(&c.count, 0)
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	atomic.AddUint64(&c.count, uint64(n))
	return n, err
}
