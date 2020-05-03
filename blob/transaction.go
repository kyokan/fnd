package blob

import (
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

var (
	ErrTransactionClosed  = errors.New("transaction already closed")
	ErrTransactionRemoved = errors.New("transaction removed")
)

type Transaction interface {
	Readable
	io.WriterAt
	WriteSector(id uint8, sector Sector) error
	Truncate() error
	Commit() error
	Rollback() error
	Remove() error
}

type txImpl struct {
	name        string
	f           *os.File
	mu          sync.Mutex
	cloner      func() (*os.File, error)
	committer   func(clone *os.File) error
	remover     func() error
	initialized bool
	closed      bool
	removed     bool
}

func (t *txImpl) Name() string {
	return t.name
}

func (t *txImpl) ReadSector(id uint8) (Sector, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ZeroSector, ErrTransactionClosed
	}
	if t.removed {
		return ZeroSector, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return ZeroSector, errors.Wrap(err, "error initializing transaction")
	}
	return ReadSector(t.f, id)
}

func (t *txImpl) ReadAt(p []byte, off int64) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, ErrTransactionClosed
	}
	if t.removed {
		return 0, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return 0, errors.Wrap(err, "error initializing transaction")
	}
	return ReadBlobAt(t.f, p, off)
}

func (t *txImpl) WriteSector(id uint8, sector Sector) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	if t.removed {
		return ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return errors.Wrap(err, "error initializing transaction")
	}
	return WriteSector(t.f, id, sector)
}

func (t *txImpl) WriteAt(p []byte, off int64) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, ErrTransactionClosed
	}
	if t.removed {
		return 0, ErrTransactionRemoved
	}
	if err := t.lazyInitialize(); err != nil {
		return 0, errors.Wrap(err, "error initializing transaction")
	}
	return WriteBlobAt(t.f, p, off)
}

func (t *txImpl) Truncate() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.removed {
		return ErrTransactionRemoved
	}
	if t.initialized {
		if err := os.Remove(t.f.Name()); err != nil {
			return err
		}
	}
	clone, err := ioutil.TempFile("", "blobtx_*")
	if err != nil {
		return errors.Wrap(err, "error creating temporary file")
	}
	if err := clone.Truncate(Size); err != nil {
		return errors.Wrap(err, "error truncating file")
	}
	t.f = clone
	t.initialized = true
	return nil
}

func (t *txImpl) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	// all the below operations need to be atomic, so panic on errors
	// until we implement a more robust recovery method
	if t.removed {
		if err := t.remover(); err != nil {
			panic(err)
		}
	} else if t.initialized {
		if err := t.committer(t.f); err != nil {
			panic(err)
		}
		if err := t.f.Close(); err != nil {
			panic(err)
		}
		if err := os.Remove(t.f.Name()); err != nil {
			panic(err)
		}
	}
	// end atomic section
	t.closed = true
	return nil
}

func (t *txImpl) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	if !t.initialized || t.removed {
		t.closed = true
		return nil
	}
	if err := t.f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(t.f.Name()); err != nil {
		panic(err)
	}
	t.closed = true
	return nil
}

func (t *txImpl) Remove() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTransactionClosed
	}
	t.removed = true
	if !t.initialized {
		return nil
	}
	if err := t.f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(t.f.Name()); err != nil {
		panic(err)
	}
	return nil
}

func (t *txImpl) lazyInitialize() error {
	if t.initialized {
		return nil
	}
	clone, err := t.cloner()
	if err != nil {
		return errors.Wrap(err, "error initializing transaction")
	}
	t.f = clone
	t.initialized = true
	return nil
}
