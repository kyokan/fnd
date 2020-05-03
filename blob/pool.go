package blob

import (
	"github.com/pkg/errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type PoolGetter func(name string) (Blob, error)

type Pool struct {
	blobs  map[string]*poolEntry
	getter PoolGetter
	mu     sync.Mutex
}

type poolEntry struct {
	blob Blob
	refs int
}

func NewPool(getter PoolGetter) *Pool {
	return &Pool{
		blobs:  make(map[string]*poolEntry),
		getter: getter,
	}
}

func (p *Pool) Get(name string) (Blob, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry := p.blobs[name]
	if entry != nil {
		entry.refs++
		return entry.blob, nil
	}
	blob, err := p.getter(name)
	if err != nil {
		return nil, err
	}
	p.blobs[name] = &poolEntry{
		blob: blob,
		refs: 1,
	}
	return blob, nil
}

func (p *Pool) Put(blob Blob) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := blob.Name()
	entry := p.blobs[name]
	if entry == nil {
		panic("putting back unknown blob")
	}
	entry.refs--
	if entry.refs > 0 {
		return nil
	}
	delete(p.blobs, name)
	return entry.blob.Close()
}
